package hello

import (
	"context"
	"net/http"

	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/http/web"
)

type HandleHelloResponse []byte

func (r HandleHelloResponse) Data() (data []byte, contentType string, err error) {
	return r, "text/plain", nil
}

func (a *App) HandleHello(_ context.Context, _ *http.Request) web.DataResponder {
	hello, err := a.business.SayHello()
	if err != nil {
		return web.ErrorResponse{Err: err}
	}

	return HandleHelloResponse(hello)
}
