package main

import (
	"io"
	"log/slog"
	"time"

	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/config"
)

type Config struct {
	// AMQPConnect - строка подключения к брокеру RabbitMQ.
	AMQPConnect string `yaml:"amqp_connect" env:"CALENDAR_AMQP_CONNECT" env-default:"amqp://guest:guest@localhost:5672/"`

	// NotifyInterval как часто опрашивать сервис календаря на предмет событий с уведомлениями.
	NotifyInterval time.Duration `yaml:"notify_interval" env:"CALENDAR_NOTIFY_INTERVAL" env-default:"60s"`

	// PurgeInterval как часто удалять старые события.
	PurgeInterval time.Duration `yaml:"purge_interval" env:"CALENDAR_PURGE_INTERVAL" env-default:"1h"`

	// PurgeOlderThan - события старше данного значения будут удалены.
	// По умолчанию - удаляем события старше 365 дней.
	PurgeOlderThan time.Duration `yaml:"purge_older_than" env:"CALENDAR_PURGE_PERIOD" env-default:"8760h"` // 365 * 24

	Log LoggerConfig `yaml:"logger" env-prefix:"CALENDAR_LOG_"`

	EventStorageType config.EventStorageType `yaml:"event_storage"    env:"CALENDAR_EVENT_STORAGE" env-default:"memory"`
	EventStoragePg   config.EventStoragePg   `yaml:"event_storage_pg"                                                   env-prefix:"CALENDAR_EVENT_STORAGE_PG_"` //nolint:lll
}

type LoggerConfig struct {
	Level slog.Level `yaml:"level" env:"LEVEL" env-default:"info"`
}

func ReadConfig(path string) (Config, error) {
	return config.ReadConfig[Config](path)
}

func ParseConfig(r io.Reader) (Config, error) {
	return config.ParseConfig[Config](r)
}
