package event

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidEventID     = errors.New("invalid event ID")
	ErrInvalidOwnerID     = errors.New("invalid owner ID")
	ErrEmptyTitle         = errors.New("empty title")
	ErrMaxTitleLen        = errors.New("title is too long")
	ErrTimeEndBeforeStart = errors.New("the EndAt is before StartAt")
)

// ID - uuid строка, идентификатор события.
type ID string

// NewID возвращает ID.
func NewID() ID {
	return ID(uuid.NewString())
}

// NewIDFromString проверяет, что строка eventID - валидный uuid.
// Возвращает ID или ошибку валидации.
func NewIDFromString(eventID string) (ID, error) {
	if err := uuid.Validate(eventID); err != nil {
		return ID(""), fmt.Errorf("%w: %w", ErrInvalidEventID, err)
	}

	return ID(eventID), nil
}

// OwnerID - uuid строка, идентификатор владельца события.
type OwnerID string

// NewOwnerID возвращает OwnerID.
func NewOwnerID() OwnerID {
	return OwnerID(uuid.NewString())
}

// NewOwnerIDFromString проверяет, что строка ownerID - валидный uuid.
// Возвращает OwnerID или ошибку валидации.
func NewOwnerIDFromString(ownerID string) (OwnerID, error) {
	if err := uuid.Validate(ownerID); err != nil {
		return OwnerID(""), fmt.Errorf("%w: %w", ErrInvalidOwnerID, err)
	}

	return OwnerID(ownerID), nil
}

// Title - заголовок события.
//   - не пустая строка
//   - не превышает MaxEventTitleLen
type Title string

const MaxEventTitleLen = 128

// NewTitle возвращает EventTitle из title, или ошибку валидации (ErrEmptyTitle, ErrMaxTitleLen).
func NewTitle(title string) (Title, error) {
	if title == "" {
		return Title(""), ErrEmptyTitle
	}

	if len(title) > MaxEventTitleLen {
		return Title(""), ErrMaxTitleLen
	}

	return Title(title), nil
}

// Event - событие, которое запланировано в промежуток [startAt, endAt).
type Event struct {
	eventID ID      // уникальный идентификатор события
	ownerID OwnerID // уникальный идентификатор пользователя, для которого назначено событие

	startAt time.Time // дата и время события
	endAt   time.Time // дата и время окончания события

	Title        Title  // заголовок
	Description  string // описание события, опционально
	NotifyBefore uint   // за сколько дней уведомлять о событии, 0 - не уведомлять
}

// NewEvent создаёт экземпляр нового события исходя из переданных параметров.
// В случае ошибки валидации параметров будет возвращена соответствующая ошибка.
func NewEvent(eventID ID, ownerID OwnerID, title Title, startAt time.Time, endAt time.Time) (Event, error) {
	if err := validateTime(startAt, endAt); err != nil {
		return Event{}, err
	}

	return Event{
		eventID: eventID,
		ownerID: ownerID,

		startAt: startAt,
		endAt:   endAt,

		Title: title,
	}, nil
}

func (e *Event) EventID() ID {
	return e.eventID
}

func (e *Event) OwnerID() OwnerID {
	return e.ownerID
}

func (e *Event) StartAt() time.Time {
	return e.startAt
}

func (e *Event) EndAt() time.Time {
	return e.endAt
}

// validateTime проверяет, что startAt перед endAt.
// Возвращает ErrTimeEndBeforeStart.
func validateTime(startAt, endAt time.Time) error {
	if !startAt.Before(endAt) {
		return ErrTimeEndBeforeStart
	}

	return nil
}
