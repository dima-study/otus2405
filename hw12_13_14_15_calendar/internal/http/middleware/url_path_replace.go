package middleware

import (
	"net/http"
	"strings"

	internalhttp "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/http"
)

// URLPathPrefixReplace - Middleware для http.Handler.
//
// Добавляет заменяет old на prefix в запросе для URL.
func URLPathPrefixReplace(old string, prefix string) internalhttp.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = strings.Replace(r.URL.Path, old, prefix, 1)

			next.ServeHTTP(w, r)
		})
	}
}
