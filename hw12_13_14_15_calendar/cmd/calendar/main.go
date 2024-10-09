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
	"syscall"

	"github.com/ilyakaznacheev/cleanenv"

	helloAPI "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/api/hello"
	helloBusiness "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/business/hello"
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
	_ = storage

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

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info(
			"start server",
			slog.String("listen", listenAddr),
		)
		serverErrors <- server.ListenAndServe()
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

		if err := server.Shutdown(ctx); err != nil {
			server.Close()
			return fmt.Errorf("could not stop server: %w", err)
		}
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
