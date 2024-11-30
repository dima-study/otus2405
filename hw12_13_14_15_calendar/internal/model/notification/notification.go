package notification

import (
	"time"

	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/event"
)

// Notification - уведомление о событии.
type Notification struct {
	EventID event.ID
	OwnerID event.OwnerID
	Title   event.Title
	Date    time.Time
}
