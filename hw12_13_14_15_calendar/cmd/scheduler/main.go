package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ilyakaznacheev/cleanenv"

	schedulerBusiness "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/business/scheduler"
	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/config"
	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/logger"
	queue "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/queue/notify/rabbit"
	memoryStorage "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/storage/event/memory"
	pgStorage "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/storage/event/pg"
)

const serviceName = "scheduler"

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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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

	logger.Info("init notifier queue")
	notifier, notifierDoneFn, err := initNotifier(logger, cfg)
	if err != nil {
		return err
	}
	defer notifierDoneFn()

	logger.Info("init app")
	schedulerBusinessApp := schedulerBusiness.NewApp(logger, notifier, storage)
	schedulerBusinessApp.PurgeOlderThan = cfg.PurgeOlderThan
	schedulerBusinessApp.NotifyInterval = cfg.NotifyInterval

	logger.Info("start app")
	schedulerBusinessApp.Schedule(ctx)

	logger.Info("register shutdown")
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("app is started")

	sig := <-shutdown
	logger.Info(
		"shutdown",
		slog.String("signal", sig.String()),
	)

	// отменяем контекст, должен завершить schedulerBusinessApp.Schedule
	cancel()

	// ждём завершения
	schedulerBusinessApp.Wait()

	return nil
}

func initStorage(cfg Config) (schedulerBusiness.EventStorage, func() error, error) {
	switch cfg.EventStorageType {
	case config.EventStorageTypeMemory:
		storage := memoryStorage.NewStorage()
		return storage, func() error { return nil }, nil
	case config.EventStorageTypePg:
		storage, err := pgStorage.NewStorage(cfg.EventStoragePg.DataSource)
		if err != nil {
			return nil, nil, err
		}

		return storage, storage.Close, nil
	default:
		return nil, nil, fmt.Errorf("storage '%s' is not supported", cfg.EventStorageType)
	}
}

func initNotifier(logger *slog.Logger, cfg Config) (schedulerBusiness.Notifier, func(), error) {
	q := queue.NewNotifyQueue(logger, cfg.AMQPConnect)
	err := q.Init()
	if err != nil {
		return nil, nil, err
	}

	return q, func() { q.Done() }, nil
}
