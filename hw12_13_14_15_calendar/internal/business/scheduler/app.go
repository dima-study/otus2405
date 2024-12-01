package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	model "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/event"
)

type Notifier interface {
	Notify(ctx context.Context, event model.Event) error
}

type EventStorage interface {
	// PurgeOldEvents удаляет события из коллекции старше чем olderThan.
	PurgeOldEvents(ctx context.Context, olderThan time.Time) error

	// QueryEventsToNotify находит все события в коллекции, по которым необходимо отправить уведомление
	// в указанный промежуток времени [from, to).
	QueryEventsToNotify(ctx context.Context, from time.Time, to time.Time) ([]model.Event, error)
}

type App struct {
	// PurgeOlderThan - сообщения старше чем PurgeOlderThan долждны быть удалены.
	PurgeOlderThan time.Duration

	// NotifyInterval как часто уведомлять.
	NotifyInterval time.Duration

	// PurgeInterval как часто удалять.
	PurgeInterval time.Duration

	logger   *slog.Logger
	notifier Notifier
	storage  EventStorage

	done chan struct{}
	mx   sync.Mutex
	wg   sync.WaitGroup
}

// NewApp создаёт новое приложение-бизнес логику для планировщика уведомлений.
func NewApp(logger *slog.Logger, notifier Notifier, storage EventStorage) *App {
	return &App{
		PurgeOlderThan: time.Hour * 24 * 365, // по умолчанию 365 дней
		NotifyInterval: time.Minute,          // по умолчанию раз в минуту
		PurgeInterval:  time.Hour,            // по умолчанию раз в час

		logger:   logger,
		notifier: notifier,
		storage:  storage,
	}
}

// Schedule запускает планировщик.
// Планировщик выполняет периодические задания. Остановить планировщик возможно отменой контекста ctx.
// При повторном запуске работающего планировщика ничего не произойдёт.
func (a *App) Schedule(ctx context.Context) {
	a.mx.Lock()
	defer a.mx.Unlock()

	// уже запущен
	if a.done != nil {
		return
	}

	a.done = make(chan struct{})

	// контролирует завершение задач уведомлений и очистки
	a.wg.Add(2)

	// Завершение работы планировщика
	go func() {
		<-ctx.Done()

		a.mx.Lock()
		defer a.mx.Unlock()

		close(a.done)
		a.done = nil
	}()

	// задачи уведомления
	go a.scheduleNotify(a.done, a.NotifyInterval)

	// задачи очистки
	go a.schedulePurgeEvents(a.done, a.PurgeInterval, a.PurgeOlderThan)
}

// Wait ждёт завершения работы планировщика.
func (a *App) Wait() {
	// ждём пока задачи завершатся
	a.wg.Wait()
}

// scheduleNotify выполняет задачу рассылки уведомлений через Notifier.
func (a *App) scheduleNotify(done chan struct{}, period time.Duration) {
	defer a.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// notify запускается периодически, вычитывает из коллекции события для которых необходимо отправить уведомление,
	// и отправляет задачу уведомления в очередь.
	notify := func() func() {
		// получить события начиная с from...
		var from time.Time

		return func() {
			l := a.logger.WithGroup("notify")

			if from.IsZero() {
				from = time.Now()
			}

			// ... по to
			to := time.Now().Add(period)
			l.DebugContext(
				ctx,
				"retrieving events to notify",
				slog.String("from", from.String()),
				slog.String("to", to.String()),
			)

			events, err := a.storage.QueryEventsToNotify(ctx, from, to)
			if err != nil {
				l.ErrorContext(ctx, "can't query events to notify", slog.String("error", err.Error()))
				return
			}

			// при следующем вызове notify будем брать следующие события.
			from = to

			l.DebugContext(
				ctx,
				"got events to notify",
				slog.Int("count", len(events)),
			)

			for _, ev := range events {
				err := a.notifier.Notify(ctx, ev)
				if err != nil {
					l.ErrorContext(
						ctx,
						"can't send notification for the event",
						slog.String("error", err.Error()),
					)
				}
			}
		}
	}()

	// Уведомляем сразу при запуске и периодически
	notify()

	t := time.NewTicker(period)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			notify()
		case <-done:
			return
		}
	}
}

// schedulePurgeEvents выполняет задачу удаления старых события.
// События удаляются сразу после запуска и периодически раз в час.
func (a *App) schedulePurgeEvents(done chan struct{}, period time.Duration, olderThan time.Duration) {
	defer a.wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	purge := func() {
		l := a.logger.WithGroup("purge")

		l.DebugContext(ctx, "purge old events")

		err := a.storage.PurgeOldEvents(ctx, time.Now().Add(-olderThan))
		if err != nil {
			l.ErrorContext(ctx, "can't purge events", slog.String("error", err.Error()))
		}
	}

	// Удаляем сразу при запуске и периодически
	purge()

	t := time.NewTicker(period)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			purge()
		case <-done:
			return
		}
	}
}
