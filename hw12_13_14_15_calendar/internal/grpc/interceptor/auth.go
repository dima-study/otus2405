package interceptor

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/grpc/auth"
)

type AuthServerOptions struct {
	UnaryInterceptor  grpc.UnaryServerInterceptor
	StreamInterceptor grpc.StreamServerInterceptor
}

// Auth возвращает пару интерсепторов для авторизации.
// Т.к. ДЗ не требует авторизации, то здесь лишь сохранении OwnerID в контексте выполнения.
func Auth(logger *slog.Logger) AuthServerOptions {
	opts := AuthServerOptions{}

	opts.UnaryInterceptor = grpc.UnaryServerInterceptor(
		func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			next := func(ctx context.Context) (any, error) {
				return handler(ctx, req)
			}

			resp, err = authOwner(ctx, logger, next)

			return resp, err
		},
	)

	opts.StreamInterceptor = grpc.StreamServerInterceptor(
		func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			next := func(ctx context.Context) (any, error) {
				return nil, handler(srv, stream)
			}

			_, err := authOwner(stream.Context(), logger, next)

			return err
		},
	)

	return opts
}

func authOwner(
	ctx context.Context,
	logger *slog.Logger,
	next func(ctx context.Context) (any, error),
) (resp any, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no auth data")
	}

	owner := md.Get("x-owner-id")
	if len(owner) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no owner token")
	}

	ownerID, err := auth.ValidateOwner(owner[0])
	if err != nil {
		logger.InfoContext(ctx, "invalid ownerID", slog.String("error", err.Error()))
		return nil, status.Error(codes.Unauthenticated, "invalid owner token")
	}

	ctx = context.WithValue(ctx, auth.AuthOwnerKey, ownerID)
	resp, err = next(ctx)

	return resp, err
}
