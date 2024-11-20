package config

import (
	"fmt"
	"io"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// ReadConfig пытается прочитать конфиг в yaml формате из файла и переменных окружения.
func ReadConfig[T any](path string) (cfg T, err error) {
	file, err := os.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("can't read file: %w", err)
	}

	cfg, err = ParseConfig[T](file)
	if err != nil {
		return cfg, fmt.Errorf("can't parse config: %w", err)
	}

	return cfg, nil
}

// ParseConfig пытается прочитать конфиг в yaml формате из r и переменных окружения.
func ParseConfig[T any](r io.Reader) (cfg T, err error) {
	err = cleanenv.ParseYAML(r, &cfg)
	if err != nil {
		return cfg, err
	}

	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
