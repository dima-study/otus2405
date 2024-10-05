package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type DataResponder interface {
	Data() (data []byte, contentType string, err error)
}

type StatusCodeResponder interface {
	StatusCode() int
}

// HandlerFunc хендлер функция для обработки запросов внутри web.Mux.
type HandlerFunc func(ctx context.Context, r *http.Request) DataResponder

// Handler выплняет web.HandlerFunc и обрабатывает результат.
// Возвращает http.HandlerFunc для использования в http.ServeMux.
func (m *Mux) Handler(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		resp := fn(ctx, r)

		err := handleResponse(ctx, resp, w)
		if err != nil {
			m.logger.Error(fmt.Sprintf("handle response: %s", err))
		}
	}
}

// handleResponse обрабатывает результат выполнения web.HandlerFunc
// и осуществляет непосредственно запись ответа клиенту.
func handleResponse(ctx context.Context, resp DataResponder, w http.ResponseWriter) error {
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return errors.New("client disconnected, do not send response")
		}
	}

	statusCode := http.StatusOK

	switch v := resp.(type) {
	case StatusCodeResponder:
		statusCode = v.StatusCode()
	case error:
		statusCode = http.StatusInternalServerError
	default:
		if resp == nil {
			statusCode = http.StatusNoContent
		}
	}

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	data, contentType, err := resp.Data()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("respose data: %w", err)
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)

	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("response write: %w", err)
	}

	return nil
}
