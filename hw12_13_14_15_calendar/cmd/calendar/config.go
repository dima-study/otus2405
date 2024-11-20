package main

import (
	"io"
	"log/slog"
	"time"

	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/config"
)

type Config struct {
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"CALENDAR_SHUTDOWN_TIMEOUT" env-default:"5s"`

	HTTP HTTPConfig   `yaml:"http"   env-prefix:"CALENDAR_HTTP_"`
	GRPC GRPCConfig   `yaml:"grpc"   env-prefix:"CALENDAR_GRPC_"`
	Log  LoggerConfig `yaml:"logger" env-prefix:"CANELDAR_LOG_"`

	EventStorageType config.EventStorageType `yaml:"event_storage"    env:"CALENDAR_EVENT_STORAGE" env-default:"memory"`
	EventStoragePg   config.EventStoragePg   `yaml:"event_storage_pg"                                                   env-prefix:"CALENDAR_EVENT_STORAGE_PG_"` //nolint:lll
}

type HTTPConfig struct {
	Host         string        `yaml:"host"          env:"HOST"          env-default:"localhost"`
	Port         string        `yaml:"port"          env:"PORT"          env-default:"8081"`
	ReadTimeout  time.Duration `yaml:"read_timeout"  env:"READ_TIMEOUT"  env-default:"5s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" env-default:"15s"`
}

type GRPCConfig struct {
	Host string `yaml:"host" env:"HOST" env-default:"localhost"`
	Port string `yaml:"port" env:"PORT" env-default:"50051"`
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
