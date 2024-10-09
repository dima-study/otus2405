package calendar

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	model "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/event"
)

type App struct {
	logger  *slog.Logger
	storage model.Storage
}

func NewApp(logger *slog.Logger, storage model.Storage) *App {
	return &App{
		logger:  logger,
		storage: storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, event model.Event) error {
	err := a.storage.AddEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("can't create event: %w", err)
	}

	return nil
}

func (a *App) UpdateEvent(ctx context.Context, event model.Event) error {
	err := a.storage.UpdateEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("can't update event: %w", err)
	}

	return nil
}

func (a *App) DeleteEvent(ctx context.Context, ownerID model.OwnerID, eventID model.ID) error {
	err := a.storage.DeleteEvent(ctx, ownerID, eventID)
	if err != nil {
		return fmt.Errorf("can't delete event: %w", err)
	}

	return nil
}

func (a *App) GetDayEvents(
	ctx context.Context,
	ownerID model.OwnerID,
	year int,
	month int,
	day int,
) ([]model.Event, error) {
	from := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	to := time.Date(year, time.Month(month), day+1, 0, 0, 0, 0, time.UTC)

	events, err := a.storage.QueryEvents(ctx, ownerID, from, to)
	if err != nil {
		return nil, fmt.Errorf("can't get day events: %w", err)
	}

	return events, nil
}

func (a *App) GetWeekEvents(
	ctx context.Context,
	ownerID model.OwnerID,
	year int,
	month int,
	day int,
) ([]model.Event, error) {
	from := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	to := time.Date(year, time.Month(month), day+7, 0, 0, 0, 0, time.UTC)

	events, err := a.storage.QueryEvents(ctx, ownerID, from, to)
	if err != nil {
		return nil, fmt.Errorf("can't get week events: %w", err)
	}

	return events, nil
}

func (a *App) GetMonthEvents(
	ctx context.Context,
	ownerID model.OwnerID,
	year int,
	month int,
) ([]model.Event, error) {
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)

	events, err := a.storage.QueryEvents(ctx, ownerID, from, to)
	if err != nil {
		return nil, fmt.Errorf("can't get month events: %w", err)
	}

	return events, nil
}