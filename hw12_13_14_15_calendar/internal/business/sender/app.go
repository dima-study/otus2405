package sender

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"

	model "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/notification"
)

type NotificationMessage interface {
	Model() (model.Notification, error)
	Done() error
}

type App struct {
	logger   *slog.Logger
	in       <-chan NotificationMessage
	w        io.Writer
	canceled chan struct{}
	wg       sync.WaitGroup
}

// NewApp создаёт новое приложение-бизнес логику для рассылки уведомлений.
func NewApp(logger *slog.Logger, in <-chan NotificationMessage, out io.Writer) *App {
	return &App{
		logger:   logger,
		in:       in,
		w:        out,
		canceled: make(chan struct{}),
	}
}

// Send запускает рассылщика уведомлений.
// Штатно остановить возможно отменой контекста.
// Также рассыльщик будет остановлен в случае закрытия входящего канала уведомлений.
// Повторный запуск предусмотрен только в случае штатного останова рассыльщика.
//
// Возвращает true, если рассыльщик был успешно запущен.
// В противном случае приложение рассыльщика не пригодно к использованию.
func (a *App) Send(ctx context.Context) bool {
	if !a.IsReady() {
		return false
	}

	a.logger.InfoContext(ctx, "start sender")

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		for {
			select {
			case <-ctx.Done():
				a.logger.InfoContext(ctx, "stop sender")
				return
			case msg, ok := <-a.in:
				if !ok {
					close(a.canceled)
					a.logger.WarnContext(ctx, "notification input channel has been closed, stop sender")
					return
				}

				a.logger.DebugContext(ctx, "got data from notification queue")

				notification, err := msg.Model()
				if err != nil {
					msg.Done()
					a.logger.ErrorContext(ctx, "can't convert message to model", slog.String("error", err.Error()))
					continue
				}

				str := fmt.Sprintf(
					"send notification for ownerID=%s eventID=%s: %s on %s\n",
					notification.OwnerID,
					notification.EventID,
					notification.Title,
					notification.Date.String(),
				)

				_, err = a.w.Write([]byte(str))
				if err != nil {
					a.logger.ErrorContext(ctx, "can't write notification", slog.String("error", err.Error()))
				}

				msg.Done()
			}
		}
	}()

	return true
}

// IsReady показывает, готово ли приложение к запуску.
func (a *App) IsReady() bool {
	select {
	case <-a.canceled:
		return false
	default:
		return true
	}
}

// Wait ждёт завершения рассыльщика.
func (a *App) Wait() {
	a.wg.Wait()
}
