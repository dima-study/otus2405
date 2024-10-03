package event

import (
	"context"
	"errors"
	"time"
)

var (
	ErrTimeIsBusy         = errors.New("time is busy")
	ErrEventAlreadyExists = errors.New("event already exists")
	ErrEventNotFound      = errors.New("event not found")
)

// Storage - интерфейс взаимодйствия с коллекцией событий.
type Storage interface {
	// AddEvent добавляет событие в коллекцию.
	AddEvent(ctx context.Context, event Event) error

	// UpdateEvent обновляет событие в коллекции.
	UpdateEvent(ctx context.Context, event Event) error

	// FindEvent находит собитие в коллекции по ownerID и eventID.
	FindEvent(ctx context.Context, ownerID OwnerID, eventID ID) (Event, error)

	// DeleteEvent удаляет событие из коллекции по ownerID и eventID.
	DeleteEvent(ctx context.Context, ownerID OwnerID, eventID ID) error

	// QueryEvents находит все события в коллекции для ownerID, которые запланированы на указанный промежуток [from, to).
	QueryEvents(ctx context.Context, ownerID OwnerID, from time.Time, to time.Time) ([]Event, error)
}
