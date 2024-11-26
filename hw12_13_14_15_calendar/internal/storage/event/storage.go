package event

import (
	"context"
	"errors"
	"time"

	model "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/event"
)

var (
	ErrTimeIsBusy         = errors.New("time is busy")
	ErrEventAlreadyExists = errors.New("event already exists")
	ErrEventNotFound      = errors.New("event not found")
)

// Storage - интерфейс взаимодйствия с коллекцией событий.
// Данный интерфейс должно поддерживать любое хранилище.
type Storage interface {
	// AddEvent добавляет событие в коллекцию.
	AddEvent(ctx context.Context, event model.Event) error

	// UpdateEvent обновляет событие в коллекции.
	UpdateEvent(ctx context.Context, event model.Event) error

	// FindEvent находит собитие в коллекции по ownerID и eventID.
	FindEvent(ctx context.Context, ownerID model.OwnerID, eventID model.ID) (model.Event, error)

	// DeleteEvent удаляет событие из коллекции по ownerID и eventID.
	DeleteEvent(ctx context.Context, ownerID model.OwnerID, eventID model.ID) error

	// QueryEvents находит все события в коллекции для ownerID, которые запланированы на указанный промежуток [from, to).
	QueryEvents(ctx context.Context, ownerID model.OwnerID, from time.Time, to time.Time) ([]model.Event, error)

	// PurgeOldEvents удаляет события из коллекции старше чем olderThan.
	PurgeOldEvents(ctx context.Context, olderThan time.Time) error

	// QueryEventsToNotify находит все события в коллекции, по которым необходимо отправить уведомление
	// в указанный промежуток времени [from, to).
	QueryEventsToNotify(ctx context.Context, from time.Time, to time.Time) ([]model.Event, error)
}
