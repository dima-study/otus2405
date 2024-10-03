// business/hello - пакет с бизнес-логикой hello.
package hello

import (
	"context"
	"log/slog"
)

type App struct {
	logger *slog.Logger
}

func NewApp(logger *slog.Logger) *App {
	return &App{
		logger: logger,
	}
}

func (a *App) SayHello() (string, error) {
	a.logger.DebugContext(context.Background(), "SayHello")

	return "Hello there!", nil
}
