package gw

import (
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// HeaderMatchers возвращает HeaderMatcherFunc для мультиплексора grpc-gateway.
func HeaderMatchers(matchers ...runtime.HeaderMatcherFunc) runtime.HeaderMatcherFunc {
	return func(key string) (string, bool) {
		ok := true
		for _, fn := range matchers {
			key, ok = fn(key)
			if !ok {
				break
			}
		}

		return key, ok
	}
}

// NoGRPCHeaders не пропускает заголовки Grpc-Metadata снаружи.
func NoGRPCHeaders(key string) (string, bool) {
	if strings.Index(key, "Grpc-Metadata") == 0 {
		return "", false
	}

	return key, true
}

// OwnerID заменяет заголовок X-Owner-ID на Grpc-Metadata-X-Owner-ID.
func OwnerID(key string) (string, bool) {
	if key == "X-Owner-ID" {
		key = "Grpc-Metadata-X-Owner-ID"
	}

	return key, true
}
