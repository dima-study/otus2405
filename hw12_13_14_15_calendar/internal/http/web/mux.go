package web

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

type RoutesAdder interface {
	AddRoutes(mux *Mux) error
}

// Mux - web-мультиплексер - надстройка над http-мультиплексером для обработки http запросов.
type Mux struct {
	logger *slog.Logger
	mux    *http.ServeMux
}

// NewMux создаёт новый web-мультиплексер.
func NewMux(logger *slog.Logger, routesAdder RoutesAdder) (*Mux, error) {
	mux := Mux{
		logger: logger,
		mux:    http.NewServeMux(),
	}

	err := routesAdder.AddRoutes(&mux)
	if err != nil {
		return nil, err
	}

	return &mux, nil
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

// Handle создаёт http.Handle для определённого пути и метода запроса.
func (m *Mux) Handle(method string, version string, path string, handlerFn HandlerFunc) error {
	if version != "" {
		var err error
		path, err = url.JoinPath("/", version, path)
		if err != nil {
			return fmt.Errorf("url.JoinPath: %w", err)
		}
	}

	m.mux.Handle(method+" "+path, m.Handler(handlerFn))

	return nil
}
