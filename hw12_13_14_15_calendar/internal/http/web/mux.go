package web

import (
	"log/slog"
	"net/http"
)

type RoutesAdder interface {
	AddRoutes(mux *Mux)
}

// Mux - web-мультиплексер - надстройка над http-мультиплексером для обработки http запросов.
type Mux struct {
	logger *slog.Logger
	mux    *http.ServeMux
}

// NewMux создаёт новый web-мультиплексер.
func NewMux(logger *slog.Logger, routesAdder RoutesAdder) *Mux {
	mux := Mux{
		logger: logger,
		mux:    http.NewServeMux(),
	}

	routesAdder.AddRoutes(&mux)

	return &mux
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

// Handle создаёт http.Handle для определённого пути и метода запроса.
func (m *Mux) Handle(method string, version string, path string, handlerFn HandlerFunc) {
	routePath := path
	if version != "" {
		routePath = "/" + version + path
	}

	m.mux.Handle(method+" "+routePath, m.Handler(handlerFn))
}
