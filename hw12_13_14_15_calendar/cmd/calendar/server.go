package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"google.golang.org/grpc"

	grpcInterceptor "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/grpc/interceptor"
)

type (
	startServerFunc func() error
	stopServerFunc  func(context.Context) error

	grpcServiceRegisterFunc func(*grpc.Server) error
)

// createHTTPServer создаёт новый HTTP-сервер и возвращает функцию старта и останова сервера.
func createHTTPServer(
	logger *slog.Logger,
	cfg Config,
	handler http.Handler,
) (startServerFunc, stopServerFunc) {
	listenAddr := net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port)

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      handler,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), cfg.Log.Level),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	start := func() error {
		logger.Info(
			"starting HTTP server",
			slog.String("listen", listenAddr),
		)

		return server.ListenAndServe()
	}

	stop := func(ctx context.Context) error {
		if err := server.Shutdown(ctx); err != nil {
			server.Close()
			return fmt.Errorf("cound not shutdown HTTP server: %w", err)
		}

		return nil
	}

	return startServerFunc(start), stopServerFunc(stop)
}

// createGRPCServer создаёт GRPC-сервер и возвращает функции старта и останова сервера.
//
// Возвращает ошибку, случившуюся при регистрации сервиса.
func createGRPCServer(
	logger *slog.Logger,
	cfg Config,
	register grpcServiceRegisterFunc,
) (startServerFunc, stopServerFunc, error) {
	grpcLogInterceptors := grpcInterceptor.LogRequest(logger.WithGroup("grpc-request"))
	grpcAuthInterceptors := grpcInterceptor.Auth(logger.WithGroup("grpc-auth"))

	server := grpc.NewServer(
		grpcLogInterceptors.UnknownServiceHandler,

		grpc.ChainUnaryInterceptor(
			grpcLogInterceptors.UnaryInterceptor,
			grpcAuthInterceptors.UnaryInterceptor,
		),

		grpc.ChainStreamInterceptor(
			grpcLogInterceptors.StreamInterceptor,
			grpcAuthInterceptors.StreamInterceptor,
		),
	)

	if err := register(server); err != nil {
		return nil, nil, err
	}

	start := func() error {
		listenAddr := net.JoinHostPort(cfg.GRPC.Host, cfg.GRPC.Port)

		lstn, err := net.Listen("tcp", listenAddr)
		if err != nil {
			return err
		}

		logger.Info(
			"start GRPC server",
			slog.String("listen", listenAddr),
		)

		return server.Serve(lstn)
	}

	stop := func(ctx context.Context) error {
		done := make(chan struct{}, 1)
		go func() {
			defer close(done)

			server.GracefulStop()
		}()

		select {
		case <-ctx.Done():
			server.Stop()
			return fmt.Errorf("could not shutdown GRPC server: %w", ctx.Err())
		case <-done:
		}

		return nil
	}

	return startServerFunc(start), stopServerFunc(stop), nil
}

// startAndShutdown - функция старта и завершения работы серверов.
//
// Будет возвращена первая ошибка при запуске сервера или все ошибки при останове.
func startAndShutdown(
	ctx context.Context,
	logger *slog.Logger,
	cfg Config,
	starters []startServerFunc,
	stoppers []stopServerFunc,
) error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverErrors := make(chan error, len(starters))
	for _, fn := range starters {
		fn := fn
		go func() {
			if err := fn(); err != nil {
				serverErrors <- err
			}
		}()
	}

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		logger.Info(
			"shutdown",
			slog.String("signal", sig.String()),
		)

		ctx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
		defer cancel()

		errsCh := make(chan error, len(stoppers))
		wg := sync.WaitGroup{}
		wg.Add(len(stoppers))

		for _, fn := range stoppers {
			fn := fn
			go func() {
				defer wg.Done()

				if err := fn(ctx); err != nil {
					errsCh <- err
				}
			}()
		}

		go func() {
			wg.Wait()
			close(errsCh)
		}()

		var err error
		for e := range errsCh {
			err = errors.Join(err, e)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
