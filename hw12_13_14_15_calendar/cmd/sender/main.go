package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ilyakaznacheev/cleanenv"

	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/business/sender"
	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/logger"
	queue "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/queue/notify/rabbit"
)

const serviceName = "sender"

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

	logger.Info("init notifier queue")
	notificationCh, notifierDoneFn, err := initNotificationQueue(ctx, logger, cfg)
	if err != nil {
		return err
	}
	defer notifierDoneFn()

	logger.Info("init app")
	senderBusinessApp := sender.NewApp(logger, notificationCh, os.Stdout)

	logger.Info("start app")
	if !senderBusinessApp.Send(ctx) {
		return errors.New("can't start sender")
	}

	logger.Info("register shutdown")
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("app is started")

	sig := <-shutdown
	logger.Info(
		"shutdown",
		slog.String("signal", sig.String()),
	)

	// отменяем контекст, должен завершить senderBusinessApp.Send
	cancel()

	// ждём завершения
	senderBusinessApp.Wait()

	go func() {
		<-time.After(cfg.ShutdownTimeout)
		logger.Error("can't stop sender in time")
		os.Exit(1)
	}()

	return err
}

func initNotificationQueue(
	ctx context.Context,
	logger *slog.Logger,
	cfg Config,
) (<-chan sender.NotificationMessage, func(), error) {
	q := queue.NewNotifyQueue(logger, cfg.AMQPConnect)
	err := q.Init()
	if err != nil {
		return nil, nil, err
	}

	ch, err := q.RegisterReceiver(ctx)
	if err != nil {
		q.Done()
		return nil, nil, err
	}

	out := make(chan sender.NotificationMessage)
	go func() {
		defer close(out)
		for m := range ch {
			select {
			case out <- &m:
			case <-ctx.Done():
				return
			}
		}
	}()

	return out, func() { q.Done() }, nil
}
