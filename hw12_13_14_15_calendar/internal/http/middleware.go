package http

import "net/http"

type Middleware func(next http.Handler) http.Handler

// ApplyMiddlewares - хелпер для оборачивания handlerFn в список Middleware.
//
// Middleware будут выполняться в порядке следования.
func ApplyMiddlewares(handlerFn http.Handler, mw ...Middleware) http.Handler {
	n := len(mw) - 1
	for i := range len(mw) {
		m := mw[n-i]
		if m != nil {
			handlerFn = m(handlerFn)
		}
	}

	return handlerFn
}
