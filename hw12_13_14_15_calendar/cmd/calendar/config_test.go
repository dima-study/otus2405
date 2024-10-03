package main

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_ParseConfig(t *testing.T) {
	tests := []struct {
		name      string
		cfg       string
		init      func()
		want      Config
		wantError bool
	}{
		{
			name: "full config",
			cfg: `
  shutdown_timeout: 1s

  http:
    port: "12345"
    host: lolo
    read_timeout: 1s
    write_timeout: 1m

  logger:
    level: debug

  event_storage: pg
  event_storage_pg:
    data_source: pg://data?source
      `,
			want: Config{
				ShutdownTimeout: time.Second,

				HTTP: HTTPConfig{
					Host:         "lolo",
					Port:         "12345",
					ReadTimeout:  time.Second,
					WriteTimeout: time.Minute,
				},
				Log: LoggerConfig{
					Level: slog.LevelDebug,
				},
				EventStorageType: "pg",
				EventStoragePg: EventStoragePg{
					DataSource: "pg://data?source",
				},
			},
			wantError: false,
		},
		{
			name: "overwrite by env",
			cfg: `
  shutdown_timeout: 15s

  http:
    port: "12345"
    host: lolo
    read_timeout: 15s
    write_timeout: 15s

  logger:
    level: debug

  event_storage: memory
  event_storage_pg:
    data_source: unknown
      `,
			init: func() {
				os.Setenv("CALENDAR_SHUTDOWN_TIMEOUT", "1s")

				os.Setenv("CALENDAR_HTTP_HOST", "some.host")
				os.Setenv("CALENDAR_HTTP_PORT", "54321")
				os.Setenv("CALENDAR_HTTP_READ_TIMEOUT", "1s")
				os.Setenv("CALENDAR_HTTP_WRITE_TIMEOUT", "1m")

				os.Setenv("CANELDAR_LOG_LEVEL", "error")

				os.Setenv("CALENDAR_EVENT_STORAGE", "pg")
				os.Setenv("CALENDAR_EVENT_STORAGE_PG_DATASOURCE", "pg://data?source")
			},
			want: Config{
				ShutdownTimeout: time.Second,

				HTTP: HTTPConfig{
					Host:         "some.host",
					Port:         "54321",
					ReadTimeout:  time.Second,
					WriteTimeout: time.Minute,
				},
				Log: LoggerConfig{
					Level: slog.LevelError,
				},
				EventStorageType: "pg",
				EventStoragePg: EventStoragePg{
					DataSource: "pg://data?source",
				},
			},
			wantError: false,
		},
		{
			name: "default",
			cfg:  `default: true`,
			want: Config{
				ShutdownTimeout: 5 * time.Second,

				HTTP: HTTPConfig{
					Host:         "localhost",
					Port:         "8081",
					ReadTimeout:  5 * time.Second,
					WriteTimeout: 15 * time.Second,
				},
				Log: LoggerConfig{
					Level: slog.LevelInfo,
				},
				EventStorageType: "memory",
				EventStoragePg:   EventStoragePg{},
			},
			wantError: false,
		},
		{
			name: "invalid log level",
			cfg: `
      http:
        port: "12345"
        host: "lolo"

      logger:
        level: failed
          `,
			wantError: true,
		},
		{
			name: "invalid log level env",
			cfg:  `default: true`,
			init: func() {
				os.Setenv("CANELDAR_LOG_LEVEL", "failed")
			},
			wantError: true,
		},
		{
			name: "invalid event storage type",
			cfg: `
		http:
		  port: "12345"
		  host: "lolo"

		logger:
		  level: debug

		event_storage: pga
		event_storage_pg:
		  data_source: pg://data?source
		    `,
			wantError: true,
		},
		{
			name: "invalid event storage env",
			cfg:  `default: true`,
			init: func() {
				os.Setenv("CALENDAR_EVENT_STORAGE", "failed")
			},
			wantError: true,
		},
	}

	for i, tt := range tests {
		name := tt.name
		if name == "" {
			name = strconv.Itoa(i)
		}

		t.Run(name, func(t *testing.T) {
			if tt.init != nil {
				tt.init()
			}

			r := strings.NewReader(tt.cfg)
			cfg, err := ParseConfig(r)

			if tt.wantError {
				require.Error(t, err, "must have error")
			} else {
				require.NoError(t, err, "hust not have error")
				require.Equal(t, tt.want, cfg, "must be equal")
			}

			os.Unsetenv("CALENDAR_SHUTDOWN_TIMEOUT")

			os.Unsetenv("CALENDAR_HTTP_PORT")
			os.Unsetenv("CALENDAR_HTTP_HOST")
			os.Unsetenv("CALENDAR_HTTP_READ_TIMEOUT")
			os.Unsetenv("CALENDAR_HTTP_WRITE_TIMEOUT")

			os.Unsetenv("CANELDAR_LOG_LEVEL")

			os.Unsetenv("CALENDAR_EVENT_STORAGE")
			os.Unsetenv("CALENDAR_EVENT_STORAGE_PG_DATASOURCE")
		})
	}
}
