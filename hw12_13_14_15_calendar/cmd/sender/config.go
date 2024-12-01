package main

import (
	"io"
	"log/slog"
	"time"

	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/config"
)

type Config struct {
	// Время ожидания принудительного завершения работы.
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"CALENDAR_SHUTDOWN_TIMEOUT" env-default:"5s"`

	// AMQPConnect - строка подключения к брокеру RabbitMQ.
	AMQPConnect string `yaml:"amqp_connect" env:"CALENDAR_AMQP_CONNECT" env-default:"amqp://guest:guest@localhost:5672/"`

	Log LoggerConfig `yaml:"logger" env-prefix:"CALENDAR_LOG_"`
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
