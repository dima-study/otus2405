package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"

	calendarAPI "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/api/calendar"
	helloAPI "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/api/hello"
	pbEventV1 "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/api/proto/event/v1"
	calendarBusiness "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/business/calendar"
	helloBusiness "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/business/hello"
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

	logger.Info("init app")

	helloBusinessApp := helloBusiness.NewApp(logger.With(slog.String("comp", "business-hello")))
	calendarBusinessApp := calendarBusiness.NewApp(logger.With(slog.String("comp", "business-calendar")), storage)

	helloAPIApp := helloAPI.NewApp(helloBusinessApp, logger.With(slog.String("comp", "api-hello")))
	calendarAPIApp := calendarAPI.NewApp(calendarBusinessApp, logger.With(slog.String("comp", "api-calendar")))

	webMux, err := web.NewMux(
		logger.With(slog.String("comp", "web-mux")),
		helloAPIApp,
	)
	if err != nil {
		return fmt.Errorf("can't create web-mux: %w", err)
	}

	httpStart, httpStop := createHTTPServer(
		logger,
		cfg,
		webMux,
	)

	grpcRegisterFn := grpcServiceRegisterFunc(func(s *grpc.Server) {
		pbEventV1.RegisterEventServiceServer(s, calendarAPIApp)
	})

	grpcStart, grpcStop := createGRPCServer(logger, cfg, grpcRegisterFn)

	return startAndShutdown(
		ctx,
		logger,
		cfg,
		[]startServerFunc{httpStart, grpcStart},
		[]stopServerFunc{httpStop, grpcStop},
	)
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
