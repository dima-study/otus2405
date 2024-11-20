package pg

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // импорт pgx.stdlib для регистрации драйвера в database/sql
	"github.com/jmoiron/sqlx"

	model "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/event"
	storage "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/storage/event"
)

type pgEvent struct {
	ID           int            `db:"id"`
	EventID      string         `db:"event_id"`
	OwnerID      string         `db:"owner_id"`
	StartAt      time.Time      `db:"start_at"`
	EndAt        time.Time      `db:"end_at"`
	Title        string         `db:"title"`
	Description  sql.NullString `db:"description"`
	NotifyBefore uint           `db:"notify_before"`
}

type Storage struct {
	DB *sqlx.DB
}

var _ storage.Storage = (*Storage)(nil)

func NewStorage(dataSource string) (*Storage, error) {
	db, err := sqlx.Connect("pgx", dataSource)
	if err != nil {
		return nil, err
	}

	return &Storage{DB: db}, nil
}

func (s *Storage) Close() error {
	return s.DB.Close()
}

func (s *Storage) AddEvent(ctx context.Context, event model.Event) (err error) {
	return s.withTx(ctx, func(tx *sqlx.Tx) error {
		ev := toPgEvent(event)
		_, err := tx.NamedExecContext(
			ctx,
			`
INSERT INTO
  events (
      event_id
    , owner_id
    , time
    , title
    , description
    , notify_before
  )
VALUES (
  :event_id
  , :owner_id
  , tsrange(:start_at, :end_at)
  , :title
  , :description
  , :notify_before
)`,
			ev,
		)
		if err != nil {
			return handleModelError(err)
		}

		return nil
	})
}

func (s *Storage) UpdateEvent(ctx context.Context, event model.Event) error {
	return s.withTx(ctx, func(tx *sqlx.Tx) error {
		ev := toPgEvent(event)
		result, err := tx.NamedExec(
			`
UPDATE events
SET
    time          = tsrange(:start_at, :end_at)
  , title         = :title
  , description   = :description
  , notify_before = :notify_before

WHERE owner_id = :owner_id
  AND event_id = :event_id`,
			ev,
		)
		if err != nil {
			return handleModelError(err)
		}

		n, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if n == 0 {
			return storage.ErrEventNotFound
		}

		return nil
	})
}

func (s *Storage) FindEvent(ctx context.Context, ownerID model.OwnerID, eventID model.ID) (model.Event, error) {
	ev := pgEvent{}
	err := s.DB.GetContext(
		ctx,
		&ev,
		`
SELECT
    id
  , event_id
  , owner_id
  , lower(time) AS start_at
  , upper(time) AS end_at
  , title
  , description
  , notify_before

FROM events

WHERE owner_id=$1
  AND event_id=$2`,
		ownerID, eventID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = storage.ErrEventNotFound
		}

		return model.Event{}, err
	}

	event, err := toModel(ev)
	if err != nil {
		return model.Event{}, err
	}

	return event, nil
}

func (s *Storage) DeleteEvent(ctx context.Context, ownerID model.OwnerID, eventID model.ID) error {
	return s.withTx(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(
			ctx,
			`
DELETE

FROM events

WHERE owner_id = $1
  AND event_id = $2`,
			ownerID, eventID,
		)
		if err != nil {
			return err
		}

		n, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if n == 0 {
			return storage.ErrEventNotFound
		}

		return nil
	})
}

func (s *Storage) QueryEvents(
	ctx context.Context,
	ownerID model.OwnerID,
	from time.Time,
	to time.Time,
) ([]model.Event, error) {
	rows, err := s.DB.QueryxContext(
		ctx,
		`
SELECT
    id
  , event_id
  , owner_id
  , lower(time) AS start_at
  , upper(time) AS end_at
  , title
  , description
  , notify_before

FROM events

WHERE owner_id = $1
  AND time && tsrange($2, $3)

ORDER BY start_at`,
		ownerID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var ev pgEvent
		if err = rows.StructScan(&ev); err != nil {
			return nil, err
		}

		event, err := toModel(ev)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// withTx выполняет функцию fn в транзакции.
func (s *Storage) withTx(ctx context.Context, fn func(tx *sqlx.Tx) error) (err error) {
	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		err = finishTx(err, tx)
	}()

	err = fn(tx)

	return err
}

// finishTx завершает транзакцию tx:
//   - откатом, если есть ошибка err
//   - фиксацией, если ошибка отсутствует.
//
// Возвращает переданную ошибку err и обёртку (wrap) с возникшей ошибкой
// во время завершения транзакции.
func finishTx(err error, tx *sqlx.Tx) error {
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}

		return err
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return commitErr
	}

	return nil
}

func toPgEvent(event model.Event) pgEvent {
	ev := pgEvent{
		EventID:      string(event.EventID()),
		OwnerID:      string(event.OwnerID()),
		StartAt:      event.StartAt(),
		EndAt:        event.EndAt(),
		Title:        string(event.Title),
		Description:  sql.NullString{},
		NotifyBefore: event.NotifyBefore,
	}

	if event.Description != "" {
		ev.Description.String = event.Description
		ev.Description.Valid = true
	}

	return ev
}

func toModel(ev pgEvent) (model.Event, error) {
	eventID, err := model.NewIDFromString(ev.EventID)
	if err != nil {
		return model.Event{}, err
	}

	ownerID, err := model.NewOwnerIDFromString(ev.OwnerID)
	if err != nil {
		return model.Event{}, err
	}

	title, err := model.NewTitle(ev.Title)
	if err != nil {
		return model.Event{}, err
	}

	event, err := model.NewEvent(eventID, ownerID, title, ev.StartAt, ev.EndAt)
	if err != nil {
		return model.Event{}, err
	}

	if ev.Description.Valid {
		event.Description = ev.Description.String
	}

	if ev.NotifyBefore > 0 {
		event.NotifyBefore = ev.NotifyBefore
	}

	return event, err
}

// handleModelError пытается по содержимому ошибки err вернуть ошибку модели, если возможно,
// в противном случае возвращает переданную ошибку err.
func handleModelError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "uniq_owner_event_id"):
		return storage.ErrEventAlreadyExists
	case strings.Contains(errStr, "no_time_overlap"):
		return storage.ErrTimeIsBusy
	}

	return err
}
