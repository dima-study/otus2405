package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	model "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/event"
	storage "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/storage/event"
)

type (
	Events []model.Event

	Storage struct {
		userMap map[model.OwnerID]Events
		mx      sync.RWMutex
	}
)

var _ storage.Storage = (*Storage)(nil)

func NewStorage() *Storage {
	return &Storage{
		userMap: map[model.OwnerID]Events{},
	}
}

func (m *Storage) AddEvent(ctx context.Context, event model.Event) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	return m.addEvent(ctx, event)
}

func (m *Storage) addEvent(ctx context.Context, event model.Event) error {
	if _, err := m.findEvent(ctx, event.OwnerID(), event.EventID()); err == nil {
		return storage.ErrEventAlreadyExists
	}

	events, exists := m.userMap[event.OwnerID()]
	if !exists {
		events = Events{}
	}

	i := findNewEventIndex(events, event)
	if i == -1 {
		return storage.ErrTimeIsBusy
	}

	events = append(events[:i], append(Events{event}, events[i:]...)...)
	m.userMap[event.OwnerID()] = events

	return nil
}

func (m *Storage) FindEvent(ctx context.Context, ownerID model.OwnerID, eventID model.ID) (model.Event, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	return m.findEvent(ctx, ownerID, eventID)
}

func (m *Storage) findEvent(_ context.Context, ownerID model.OwnerID, eventID model.ID) (model.Event, error) {
	events, exists := m.userMap[ownerID]
	if !exists {
		return model.Event{}, storage.ErrEventNotFound
	}

	i := findEventIndex(events, eventID)
	if i == -1 {
		return model.Event{}, storage.ErrEventNotFound
	}

	return events[i], nil
}

func (m *Storage) UpdateEvent(ctx context.Context, event model.Event) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	return m.updateEvent(ctx, event)
}

func (m *Storage) updateEvent(ctx context.Context, event model.Event) error {
	oldEvent, err := m.findEvent(ctx, event.OwnerID(), event.EventID())
	if err != nil {
		return err
	}

	err = m.deleteEvent(ctx, oldEvent.OwnerID(), oldEvent.EventID())
	if err != nil {
		return fmt.Errorf("can't delete old event while update: %w", err)
	}

	if err := m.addEvent(ctx, event); err != nil {
		err = fmt.Errorf("can't add updated event: %w", err)

		if revertErr := m.addEvent(ctx, oldEvent); revertErr != nil {
			revertErr = fmt.Errorf("can't revert updated event: %w", revertErr)
			err = errors.Join(err, revertErr)
		}

		return err
	}

	return nil
}

func (m *Storage) DeleteEvent(ctx context.Context, ownerID model.OwnerID, eventID model.ID) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	return m.deleteEvent(ctx, ownerID, eventID)
}

func (m *Storage) deleteEvent(_ context.Context, ownerID model.OwnerID, eventID model.ID) error {
	events, exists := m.userMap[ownerID]
	if !exists {
		return storage.ErrEventNotFound
	}

	i := findEventIndex(events, eventID)
	if i == -1 {
		return storage.ErrEventNotFound
	}

	events = append(events[:i], events[i+1:]...)
	m.userMap[ownerID] = events

	return nil
}

func (m *Storage) QueryEvents(
	_ context.Context,
	ownerID model.OwnerID,
	from time.Time,
	to time.Time,
) ([]model.Event, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	events, exists := m.userMap[ownerID]

	if !exists {
		return nil, nil
	}

	k, l := -1, -1
	for i := range len(events) {
		if events[i].EndAt().Before(from) || events[i].EndAt().Equal(from) ||
			events[i].StartAt().After(to) || events[i].StartAt().Equal(to) {
			continue
		}

		if k == -1 {
			k = i
		} else {
			l = i
		}
	}

	if k == -1 {
		return nil, nil
	}

	if l < k {
		l = k
	}

	return events[k : l+1], nil
}

func (m *Storage) PurgeOldEvents(ctx context.Context, olderThan time.Time) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	return m.purgeOldEvents(ctx, olderThan)
}

func (m *Storage) purgeOldEvents(_ context.Context, olderThan time.Time) error {
	for ownerID, events := range m.userMap {
		l := len(events)
		for i := l - 1; i >= 0; i-- {
			println(i, len(events))
			if events[i].EndAt().Before(olderThan) {
				events = append(events[:i], events[i+1:]...)
				l--
			}
		}

		m.userMap[ownerID] = events[0:l:l]
	}

	return nil
}

func (m *Storage) QueryEventsToNotify(
	_ context.Context,
	from time.Time,
	to time.Time,
) ([]model.Event, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	var events []model.Event

	for _, userEvents := range m.userMap {
		for _, ev := range userEvents {
			if ev.NotifyBefore == 0 {
				continue
			}

			startAt := ev.StartAt().Add(time.Duration(-ev.NotifyBefore) * 24 * time.Hour)
			if (from.Before(startAt) || from.Equal(startAt)) && startAt.Before(to) {
				events = append(events, ev)
			}
		}
	}

	return events, nil
}

// findNewEventIndex пытается найти индекс в слайсе events для нового события event.
// Все элементы слайса с найденным индексом и выше должны располагаться "правее" элемента event
// после его добавления слайс.
// Возвращает -1, если событие event не может быть добавлено в слайс events:
// это происходит, когда время события event пересекается со временем событий в слайсе events.
func findNewEventIndex(events Events, event model.Event) int {
	if len(events) == 0 {
		return 0
	}

	k := -1
	for i := range len(events) {
		if events[i].StartAt().After(event.StartAt()) {
			k = i
			break
		}
	}

	if k == -1 {
		k = len(events)

		if events[k-1].EndAt().Before(event.StartAt()) || events[k-1].EndAt().Equal(event.StartAt()) {
			return k
		}

		return -1
	}

	if k == 0 {
		if events[k].StartAt().After(event.EndAt()) || events[k].StartAt().Equal(event.EndAt()) {
			return k
		}

		return -1
	}

	if (events[k-1].EndAt().Before(event.StartAt()) || events[k-1].EndAt().Equal(event.StartAt())) &&
		(events[k].StartAt().After(event.EndAt()) || events[k].StartAt().Equal(event.EndAt())) {
		return k
	}

	return -1
}

// findEventIndex пытается найти индекс элемента в слайсе events для указанного eventID.
// Возращает -1, если событие с eventID не найдено в слайсе events.
func findEventIndex(events Events, eventID model.ID) int {
	for i := range len(events) {
		if events[i].EventID() == eventID {
			return i
		}
	}

	return -1
}
