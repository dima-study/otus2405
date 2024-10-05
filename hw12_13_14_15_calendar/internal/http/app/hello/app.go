// http/app/hello - приложение с контроллером hello.
// Представляет собой пакет с обработчиками для http-сервера основанного на http/web/mux мультиплексоре.
//
// Основная задача - связь бизнес-логики business/hello, запросов через http сервер и формирование ответов.
package hello

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/http/web"
)

type Business interface {
	SayHello() (string, error)
}

type App struct {
	business Business
	logger   *slog.Logger
}

func NewApp(business Business, logger *slog.Logger) *App {
	return &App{
		business: business,
		logger:   logger,
	}
}

func (a *App) AddRoutes(mux *web.Mux) error {
	a.logger.Debug("add routes")

	err := mux.Handle(http.MethodGet, "", "/hello", a.HandleHello)
	if err != nil {
		return fmt.Errorf("mux.Handle method=%s version=%s path=%s", http.MethodGet, "", "/hello")
	}

	return nil
}
