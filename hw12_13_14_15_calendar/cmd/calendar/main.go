package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"

	calendarAPI "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/api/calendar"
	helloAPI "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/api/hello"
	pbEventV1 "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/api/proto/event/v1"
	calendarBusiness "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/business/calendar"
	helloBusiness "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/business/hello"
	grpcInterceptor "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/grpc/interceptor"
	internalhttp "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/http"
	httpMiddleware "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/http/middleware"
	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/http/web"
	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/logger"
	model "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/event"
	memoryStorage "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/storage/event/memory"
	pgStorage "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/storage/event/pg"
)

const serviceName = "calendar"

func main() {
	initFlag()

	logger, levelVar := logger.New(os.Stdout, slog.LevelInfo, serviceName)

	ctx := context.Background()
	if err := run(ctx, logger, levelVar); err != nil {
		logger.Error(
			"failed to run",
			slog.String("error", err.Error()),
		)
	}
}

var configFile string

func initFlag() {
	flag.StringVar(&configFile, "config", "", "Path to configuration file")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()

		var cfg Config
		help, _ := cleanenv.GetDescription(&cfg, nil)
		fmt.Fprintf(flag.CommandLine.Output(), "\n%s\n", help)
	}

	flag.Parse()
}

func run(ctx context.Context, logger *slog.Logger, levelVar *slog.LevelVar) error {
	logger.Info(
		"starting service",
		slog.Group(
			"build",
			slog.String("release", release),
			slog.String("date", buildDate),
			slog.String("gitHash", gitHash),
		),
	)

	logger.Info(
		"read config",
		slog.String("file", configFile),
	)
	cfg, err := ReadConfig(configFile)
	if err != nil {
		return err
	}

	logger.Info(
		"set logger level",
		slog.String("from", levelVar.Level().String()),
		slog.String("to", cfg.Log.Level.String()),
	)
	levelVar.Set(cfg.Log.Level)

	logger.Info(
		"init storage",
		slog.String("storage", cfg.EventStorageType.String()),
	)
	storage, storageDoneFn, err := initStorage(cfg)
	if err != nil {
		return err
	}
	defer storageDoneFn()

	calendarBusinessApp := calendarBusiness.NewApp(logger, storage)

	helloBusinessApp := helloBusiness.NewApp(logger)
	helloAPIApp := helloAPI.NewApp(helloBusinessApp, logger)

	webMux, err := web.NewMux(
		logger,
		helloAPIApp,
	)
	if err != nil {
		return err
	}

	listenAddr := net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port)
	server := http.Server{
		Addr: listenAddr,
		Handler: internalhttp.ApplyMiddlewares(
			webMux,
			httpMiddleware.LogRequest(logger.WithGroup("http-request")),
		),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), cfg.Log.Level),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	grpcAuthInterceptors := grpcInterceptor.Auth(logger.WithGroup("grpc-auth"))

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcAuthInterceptors.UnaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			grpcAuthInterceptors.StreamInterceptor,
		),
	)
	grpcLstn, err := net.Listen("tcp", ":12345")
	if err != nil {
		return err
	}

	pbEventV1.RegisterEventServiceServer(grpcServer, calendarAPI.NewApp(calendarBusinessApp, logger))

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverErrors := make(chan error, 2)
	go func() {
		logger.Info(
			"start http server",
			slog.String("listen", listenAddr),
		)
		serverErrors <- server.ListenAndServe()
	}()

	go func() {
		logger.Info(
			"start grpc server",
			slog.String("listen", listenAddr),
		)
		serverErrors <- grpcServer.Serve(grpcLstn)
	}()

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

		errsCh := make(chan error, 2)
		wg := sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()

			if err := server.Shutdown(ctx); err != nil {
				server.Close()
				errsCh <- fmt.Errorf("could not stop http server: %w", err)
			}
		}()

		go func() {
			defer wg.Done()

			done := make(chan struct{}, 1)
			go func() {
				grpcServer.GracefulStop()
				done <- struct{}{}
			}()

			select {
			case <-ctx.Done():
				grpcServer.Stop()
				errsCh <- fmt.Errorf("could not stop server: %w", ctx.Err())
			case <-done:
			}
		}()

		wg.Wait()
	}

	return nil
}

func initStorage(cfg Config) (model.Storage, func() error, error) {
	switch cfg.EventStorageType {
	case EventStorageTypeMemory:
		storage := memoryStorage.NewStorage()
		return storage, func() error { return nil }, nil
	case EventStorageTypePg:
		storage, err := pgStorage.NewStorage(cfg.EventStoragePg.DataSource)
		if err != nil {
			return nil, nil, err
		}

		return storage, storage.Close, nil
	default:
		return nil, nil, fmt.Errorf("storage '%s' is not supported", cfg.EventStorageType)
	}
}
