// api/calendar - приложение с контроллером calendar.
// Представляет собой пакет с обработчиками для grpc-серера EventServiceServer
//
// Основная задача - связь бизнес-логики business/calendar, запросов через grpc сервер и формирование ответов.
package calendar

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	proto "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/api/proto/event/v1"
	model "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/event"
	storage "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/storage/event"
)

type Business interface {
	CreateEvent(ctx context.Context, event model.Event) error
	FindEvent(ctx context.Context, ownerID model.OwnerID, eventID model.ID) (model.Event, error)
	UpdateEvent(ctx context.Context, event model.Event) error
	DeleteEvent(ctx context.Context, ownerID model.OwnerID, eventID model.ID) error
	GetDayEvents(ctx context.Context, ownerID model.OwnerID, year int, month int, day int) ([]model.Event, error)
	GetWeekEvents(ctx context.Context, ownerID model.OwnerID, year int, month int, day int) ([]model.Event, error)
	GetMonthEvents(ctx context.Context, ownerID model.OwnerID, year int, month int) ([]model.Event, error)
}

type App struct {
	business Business
	logger   *slog.Logger

	proto.UnimplementedEventServiceServer
}

func NewApp(business Business, logger *slog.Logger) *App {
	return &App{
		business: business,
		logger:   logger,
	}
}

// handleError проверяет, если ошибка - не ошибка модели, то добавляет ошибку в лог.
// Возвращает grpc-ошибку со статусом.
func (a *App) handleError(ctx context.Context, err error, handle string, attrs ...any) error {
	switch {
	case errors.Is(err, model.ErrInvalidEventID):
	case errors.Is(err, model.ErrInvalidOwnerID):
	case errors.Is(err, model.ErrEmptyTitle):
	case errors.Is(err, model.ErrMaxTitleLen):
	case errors.Is(err, model.ErrTimeEndBeforeStart):
	case errors.Is(err, storage.ErrTimeIsBusy):
	case errors.Is(err, storage.ErrEventAlreadyExists):
	case errors.Is(err, storage.ErrEventNotFound):
	default:
		a.logger.
			With(append([]any{slog.String("handle", handle)}, attrs...)...).
			ErrorContext(ctx, "error occurred", slog.String("error", err.Error()))

		return status.Error(codes.Internal, "some error")
	}

	return status.Error(codes.InvalidArgument, err.Error())
}

func whereAttr(where string) slog.Attr {
	return slog.String("where", where)
}
