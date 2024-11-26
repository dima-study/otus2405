package rabbit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/event"
	model "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/notification"
)

const (
	exchangeName = "calendar"
	queueName    = "calendar.scheduler"
	routeKey     = "notify"
)

var ErrNotInitialized = errors.New("not initialized")

// Notification объект уведомления в очереди rabbitmq.
type Notification struct {
	EventID string    `json:"eventId"`
	OwnerID string    `json:"ownerId"`
	Title   string    `json:"title"`
	Date    time.Time `json:"startAt"`

	m *amqp.Delivery `json:"-"`
}

func NewNotification(e event.Event) Notification {
	return Notification{
		EventID: string(e.EventID()),
		OwnerID: string(e.OwnerID()),
		Title:   string(e.Title),
		Date:    e.StartAt(),
	}
}

// Done подтверждает, что уведомление было обработано после чтения из очереди.
// Если сообщение не было обработано по какой-либо причине, то оно вернётся обратно в очередь для повторной обработки.
func (n *Notification) Done() error {
	return n.m.Ack(false)
}

// Model возвращает модель уведомления, если возможно.
func (n *Notification) Model() (model.Notification, error) {
	eventID, err := event.NewIDFromString(n.EventID)
	if err != nil {
		return model.Notification{}, err
	}

	ownerID, err := event.NewOwnerIDFromString(n.OwnerID)
	if err != nil {
		return model.Notification{}, err
	}

	title, err := event.NewTitle(n.Title)
	if err != nil {
		return model.Notification{}, err
	}

	return model.Notification{
		EventID: eventID,
		OwnerID: ownerID,
		Title:   title,
		Date:    n.Date,
	}, nil
}

func (n *Notification) Marshal() ([]byte, error) {
	b, err := json.Marshal(n)
	if err != nil {
		return nil, err
	}

	return b, err
}

func (n *Notification) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, n)
	if err != nil {
		return err
	}

	_, err = n.Model()
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

type NotifyQueue struct {
	connectString string
	logger        *slog.Logger

	conn *amqp.Connection
	ch   *amqp.Channel
	mx   sync.Mutex
}

func NewNotifyQueue(logger *slog.Logger, connect string) *NotifyQueue {
	return &NotifyQueue{
		connectString: connect,
		logger:        logger,
	}
}

// Init инициализирует очередь rabbitmq.
func (q *NotifyQueue) Init() (err error) {
	q.mx.Lock()

	defer func() {
		q.mx.Unlock()

		if err != nil {
			q.Done()
		}
	}()

	// 1 подключаемся
	conn, err := amqp.Dial(q.connectString)
	if err != nil {
		return fmt.Errorf("can't create connection: %w", err)
	}
	q.conn = conn

	// 2 новый канал связи
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("can't create channel: %w", err)
	}
	q.ch = ch

	// 3 определяем exchange
	err = ch.ExchangeDeclare(
		exchangeName,
		"direct",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("can't create exchange: %w", err)
	}

	// 4 определяем queue
	queue, err := q.ch.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("can't create exchange: %w", err)
	}

	// 5 связываем queue и exchange через routeKey
	err = q.ch.QueueBind(
		queue.Name,
		routeKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("can't bind queue to exchange: %w", err)
	}

	return nil
}

// RegisterReceiver регистрирует новый "приёмник" для сообщений из очереди уведомлений.
// Возвращает канал с уведомлениями или ошибку.
// Завершить работу приёмника возможно отменой контекста.
func (q *NotifyQueue) RegisterReceiver(ctx context.Context) (<-chan Notification, error) {
	q.mx.Lock()
	defer q.mx.Unlock()

	if q.ch == nil {
		return nil, ErrNotInitialized
	}

	consumerID := uuid.New()

	msgs, err := q.ch.ConsumeWithContext(
		ctx,
		queueName,
		consumerID.String(),
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("can't consume from the queue: %w", err)
	}

	out := make(chan Notification)
	go func() {
		defer close(out)

		for m := range msgs {
			if ctx.Err() != nil {
				return
			}

			var n Notification
			if err := n.Unmarshal(m.Body); err != nil {
				q.logger.Error("can't decode notification from queue", slog.String("error", err.Error()))
				m.Ack(false)
				continue
			}

			n.m = &m

			out <- n
		}
	}()

	return out, nil
}

// Notify отправляет уведомление по событию event в очередь.
// Возвращает ошибку, если отправить не удалось.
func (q *NotifyQueue) Notify(ctx context.Context, event event.Event) error {
	q.mx.Lock()
	defer q.mx.Unlock()

	if q.ch == nil {
		return ErrNotInitialized
	}

	n := NewNotification(event)
	data, err := n.Marshal()
	if err != nil {
		return fmt.Errorf("can't marshal notification: %w", err)
	}

	err = q.ch.PublishWithContext(
		ctx,
		exchangeName,
		routeKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)
	if err != nil {
		return fmt.Errorf("can't publish notification: %w", err)
	}

	return nil
}

// Done завершает работу с очередью и закрывает соединение.
func (q *NotifyQueue) Done() {
	q.mx.Lock()
	defer q.mx.Unlock()

	if q.ch != nil {
		q.ch.Close()
		q.ch = nil
	}

	if q.conn != nil {
		q.conn.Close()
		q.conn = nil
	}
}
