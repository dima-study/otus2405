package interceptor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type LogRequestServerOptions struct {
	UnaryInterceptor      grpc.UnaryServerInterceptor
	StreamInterceptor     grpc.StreamServerInterceptor
	UnknownServiceHandler grpc.ServerOption
}

// LogRequest возвращает пару интерсепторов для логирования запросов (аналогично HTTP логгеру).
// Также возвращает хендлер для неизвестных методов-сервисов (для логирования).
//
// NOTE: вообще можно использовать grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging
// для логирования.
// Здесь своя реализация, т.к. того требует ДЗ.
func LogRequest(logger *slog.Logger) LogRequestServerOptions {
	opts := LogRequestServerOptions{}

	opts.UnaryInterceptor = grpc.UnaryServerInterceptor(
		func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			next := func() (any, error) {
				return handler(ctx, req)
			}

			resp, err = logRequest(ctx, logger, info.FullMethod, next)

			return resp, err
		},
	)

	opts.StreamInterceptor = grpc.StreamServerInterceptor(
		func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			next := func() (any, error) {
				return nil, handler(srv, stream)
			}

			_, err := logRequest(stream.Context(), logger, info.FullMethod, next)

			return err
		},
	)

	opts.UnknownServiceHandler = grpc.UnknownServiceHandler(func(srv any, stream grpc.ServerStream) error {
		method, _ := grpc.MethodFromServerStream(stream)
		return status.Error(codes.Unimplemented, fmt.Sprintf("uknown method %s", method))
	})

	return opts
}

func logRequest(
	ctx context.Context,
	logger *slog.Logger,
	method string,
	next func() (any, error),
) (resp any, err error) {
	startAt := time.Now()

	const unknown = "UNKNOWN"

	remoteAddr := unknown
	if p, ok := peer.FromContext(ctx); ok {
		remoteAddr = p.Addr.String()
	}

	userAgent := unknown
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		ua := md.Get("User-Agent")
		if len(ua) > 0 {
			userAgent = ua[0]
		}
	}

	resp, err = next()

	statusCode := codes.OK
	if err != nil {
		statusCode = status.Code(err)
	}

	logger.Info(
		"grpc request completed",
		slog.String("statusCode", statusCode.String()),
		slog.String("remoteAddr", remoteAddr),
		slog.String("method", method),
		slog.String("userAgent", userAgent),
		slog.Duration("duration", time.Since(startAt)),
	)

	return resp, err
}
