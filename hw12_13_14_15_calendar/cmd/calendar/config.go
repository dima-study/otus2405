package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"CALENDAR_SHUTDOWN_TIMEOUT" env-default:"5s"`

	HTTP HTTPConfig   `yaml:"http"   env-prefix:"CALENDAR_HTTP_"`
	GRPC GRPCConfig   `yaml:"grpc"   env-prefix:"CALENDAR_GRPC_"`
	Log  LoggerConfig `yaml:"logger" env-prefix:"CANELDAR_LOG_"`

	EventStorageType EventStorageType `yaml:"event_storage"    env:"CALENDAR_EVENT_STORAGE" env-default:"memory"`
	EventStoragePg   EventStoragePg   `yaml:"event_storage_pg"                                                   env-prefix:"CALENDAR_EVENT_STORAGE_PG_"` //nolint:lll
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

type EventStorageType string

var (
	EventStorageTypeMemory EventStorageType = "memory"
	EventStorageTypePg     EventStorageType = "pg"
)

func (t *EventStorageType) UnmarshalText(s []byte) error {
	switch string(s) {
	case string(EventStorageTypeMemory):
		*t = EventStorageTypeMemory
	case string(EventStorageTypePg):
		*t = EventStorageTypePg
	default:
		return fmt.Errorf("invalid event storage type '%s'", s)
	}

	return nil
}

func (t *EventStorageType) String() string {
	return string(*t)
}

type EventStoragePg struct {
	DataSource string `yaml:"data_source" env:"DATASOURCE"`
}

// ReadConfig пытается прочитать конфиг в yaml формате из файла и переменных окружения.
func ReadConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("can't read file: %w", err)
	}

	cfg, err := ParseConfig(file)
	if err != nil {
		return Config{}, fmt.Errorf("can't parse config: %w", err)
	}

	return cfg, nil
}

// ParseConfig пытается прочитать конфиг в yaml формате из r и переменных окружения.
func ParseConfig(r io.Reader) (Config, error) {
	var cfg Config
	err := cleanenv.ParseYAML(r, &cfg)
	if err != nil {
		return Config{}, err
	}

	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
