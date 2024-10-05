package middleware

import (
	"log/slog"
	"net/http"
	"time"

	internalhttp "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/http"
)

// statusLogger реализует http.ResponseWriter, сохраняя у себя статус-код ответа.
type statusLogger struct {
	http.ResponseWriter
	statusCode int
}

// Unwrap нужен для корректной реализации поддержки в http.ResponseController
// Задача - вернуть оригинальный http.ResponseWriter.
func (s *statusLogger) Unwrap() http.ResponseWriter {
	return s.ResponseWriter
}

func (s *statusLogger) WriteHeader(statusCode int) {
	s.statusCode = statusCode

	s.ResponseWriter.WriteHeader(statusCode)
}

// LogRequest - Middleware для http.Handler.
//
// Добавляет Info-лог в logger информацию об исполненном запросе.
func LogRequest(logger *slog.Logger) internalhttp.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextFn := func() int {
				sl := &statusLogger{ResponseWriter: w}
				next.ServeHTTP(sl, r)

				return sl.statusCode
			}

			logRequest(
				logger,
				r.RemoteAddr,
				r.Method,
				r.URL.Path,
				r.URL.RawQuery,
				r.Proto,
				r.UserAgent(),
				nextFn,
			)
		})
	}
}

func logRequest(
	logger *slog.Logger,
	remoteAddr string,
	method string,
	path string,
	query string,
	proto string,
	userAgent string,
	next func() int,
) {
	if query != "" {
		path = path + "?" + query
	}
	startAt := time.Now()

	statusCode := next()

	logger.Info(
		"request completed",
		slog.Int("statusCode", statusCode),
		slog.String("remoteAddr", remoteAddr),
		slog.String("method", method),
		slog.String("path", path),
		slog.String("proto", proto),
		slog.String("userAgent", userAgent),
		slog.Duration("duration", time.Since(startAt)),
	)
}
