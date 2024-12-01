//go:build integration

package tests

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/api/proto/event/v1"
)

func TestGRPC(t *testing.T) {
	c, err := NewGRPCClient(net.JoinHostPort(os.Getenv("CALENDAR_GRPC_HOST"), os.Getenv("CALENDAR_GRPC_PORT")))
	require.NoError(t, err, "grpc client must be created")
	s := &IntegrationSuite{
		Client:   c,
		Timeout:  5 * time.Second,
		Filename: "grpc.calendar.test",
	}
	suite.Run(t, s)
}

func TestHTTPS(t *testing.T) {
	c, err := NewHTTPClient(net.JoinHostPort(os.Getenv("CALENDAR_HTTP_HOST"), os.Getenv("CALENDAR_HTTP_PORT")))
	require.NoError(t, err, "http client must be created")

	s := &IntegrationSuite{
		Client:   c,
		Timeout:  5 * time.Second,
		Filename: "http.calendar.test",
	}
	suite.Run(t, s)
}

type Client interface {
	Done() error
	CreateEvent(ctx context.Context, ownerID string, ev *v1.Event) (*v1.Event, error)
	GetDayEvents(ctx context.Context, ownerID string, req *v1.GetDayEventsRequest) ([]*v1.Event, error)
	GetWeekEvents(ctx context.Context, ownerID string, req *v1.GetWeekEventsRequest) ([]*v1.Event, error)
	GetMonthEvents(ctx context.Context, ownerID string, req *v1.GetMonthEventsRequest) ([]*v1.Event, error)
}

type IntegrationSuite struct {
	suite.Suite

	Client   Client
	Timeout  time.Duration
	Filename string
}

func (s *IntegrationSuite) TearDownSuite() {
	if s.Client != nil {
		err := s.Client.Done()
		s.Require().NoError(err, "client must be done")
	}
}

func (s *IntegrationSuite) Test_CreateEvent() {
	ownerID := []string{
		uuid.NewString(),
		uuid.NewString(),
	}

	now := time.Now().UTC()

	tests := []struct {
		name      string
		event     *v1.Event
		ownerID   string
		wantError bool
	}{
		{
			name: "event created",
			event: &v1.Event{
				EventID:     uuid.NewString(),
				StartAt:     timestamppb.New(now),
				EndAt:       timestamppb.New(now.Add(time.Hour)),
				Title:       "event",
				Description: "some event",
			},
			ownerID:   ownerID[0],
			wantError: false,
		},
		{
			name: "event not created: overlap time",
			event: &v1.Event{
				EventID:     uuid.NewString(),
				StartAt:     timestamppb.New(now.Add(30 * time.Minute)),
				EndAt:       timestamppb.New(now.Add(time.Hour)),
				Title:       "event",
				Description: "some event",
			},
			ownerID:   ownerID[0],
			wantError: true,
		},
		{
			name: "event created: overlap time, but another user",
			event: &v1.Event{
				EventID:     uuid.NewString(),
				StartAt:     timestamppb.New(now.Add(30 * time.Minute)),
				EndAt:       timestamppb.New(now.Add(time.Hour)),
				Title:       "event",
				Description: "some event",
			},
			ownerID:   ownerID[1],
			wantError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
			defer cancel()

			event, err := s.Client.CreateEvent(ctx, tt.ownerID, tt.event)

			if tt.wantError {
				s.Require().Error(err, "client.CreateEvent must return error")
			} else {
				s.Require().NoError(err, "client.CreateEvent must not return error")
				s.Require().Equal(tt.event.EventID, event.EventID, "event IDs must be equal")
			}
		})
	}
}

func (s *IntegrationSuite) Test_GetDayEvents() {
	ownerID := []string{
		uuid.NewString(),
		uuid.NewString(),
	}

	tomorrow := time.Now().UTC().Add(24 * time.Hour) // tomorrow

	ev := &v1.Event{
		EventID: uuid.NewString(),
		StartAt: timestamppb.New(tomorrow),
		EndAt:   timestamppb.New(tomorrow.Add(24 * time.Hour)),
		Title:   "tomorrow event",
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	_, err := s.Client.CreateEvent(ctx, ownerID[1], ev)
	s.Require().NoError(err, "tomorrow event must be created")

	tests := []struct {
		name     string
		ownerID  string
		eventIDs []string
	}{
		{
			name:     "created event",
			ownerID:  ownerID[1],
			eventIDs: []string{ev.EventID},
		},
		{
			name:     "no events",
			ownerID:  ownerID[0],
			eventIDs: nil,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
			defer cancel()

			events, err := s.Client.GetDayEvents(
				ctx,
				tt.ownerID,
				&v1.GetDayEventsRequest{
					Day: &v1.Date{
						Year:  int32(tomorrow.Year()),
						Month: int32(tomorrow.Month()),
						Day:   int32(tomorrow.Day()),
					},
				},
			)
			s.Require().NoError(err, "client.GetDayEvents must not have error")

			if tt.eventIDs == nil {
				s.Require().Len(events, 0, "client.GetDayEvents must return no events")
			} else {
				eventIDs := []string{}
				for _, ev := range events {
					eventIDs = append(eventIDs, ev.EventID)
				}

				s.Require().Equal(tt.eventIDs, eventIDs, "client.GetDayEvents proper events")
			}
		})
	}
}

func (s *IntegrationSuite) Test_GetWeekEvents() {
	ownerID := []string{
		uuid.NewString(),
		uuid.NewString(),
	}

	now := time.Now().UTC()
	nextWeekStart := now.Add(3 * 24 * time.Hour) // 3 days from now
	// [
	//   1: empty,
	//   2: empty,
	//   3: empty,
	//   4: empty,
	//   5: empty 7,
	//   6: event 8,
	//   7: empty 9,
	//   8: empty 10,  - will be skipped
	// ]

	eventIDs := []string{}
	for i := range 4 {
		nextWeek := now.Add(time.Duration(7+i) * 24 * time.Hour) // create event at 7 days from now + i day
		ev := &v1.Event{
			EventID: uuid.NewString(),
			StartAt: timestamppb.New(nextWeek),
			EndAt:   timestamppb.New(nextWeek.Add(12 * time.Hour)),
			Title:   fmt.Sprintf("event %d", 7+i),
		}

		func() {
			ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
			defer cancel()

			_, err := s.Client.CreateEvent(ctx, ownerID[1], ev)
			s.Require().NoError(err, "next week event must be created")

			eventIDs = append(eventIDs, ev.EventID)
		}()
	}

	tests := []struct {
		name     string
		ownerID  string
		eventIDs []string
	}{
		{
			name:     "created events except the last one",
			ownerID:  ownerID[1],
			eventIDs: eventIDs[:len(eventIDs)-1], // except last one
		},
		{
			name:     "no events",
			ownerID:  ownerID[0],
			eventIDs: nil,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
			defer cancel()

			events, err := s.Client.GetWeekEvents(
				ctx,
				tt.ownerID,
				&v1.GetWeekEventsRequest{
					StartDay: &v1.Date{
						Year:  int32(nextWeekStart.Year()),
						Month: int32(nextWeekStart.Month()),
						Day:   int32(nextWeekStart.Day()),
					},
				},
			)
			s.Require().NoError(err, "client.GetWeekEvents must not have error")

			if tt.eventIDs == nil {
				s.Require().Len(events, 0, "client.GetWeekEvents must return no events")
			} else {
				eventIDs := []string{}
				for _, ev := range events {
					eventIDs = append(eventIDs, ev.EventID)
				}

				s.Require().Equal(tt.eventIDs, eventIDs, "client.GetWeekEvents proper events")
			}
		})
	}
}

func (s *IntegrationSuite) Test_GetMonthEvents() {
	ownerID := uuid.NewString()

	now := time.Now().UTC()

	startAt := time.Date(
		now.Year()+1,
		now.Month(),
		now.Day(),
		now.Hour(),
		now.Minute(),
		now.Second(),
		now.Nanosecond(),
		time.UTC,
	)

	ev := &v1.Event{
		EventID: uuid.NewString(),
		StartAt: timestamppb.New(startAt),
		EndAt:   timestamppb.New(startAt.Add(12 * time.Hour)),
		Title:   "next year event",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*s.Timeout)
	defer cancel()

	_, err := s.Client.CreateEvent(ctx, ownerID, ev)
	s.Require().NoError(err, "next year event must be created")

	events, err := s.Client.GetMonthEvents(
		ctx,
		ownerID,
		&v1.GetMonthEventsRequest{
			Month: &v1.Month{
				Year:  int32(startAt.Year()),
				Month: int32(startAt.Month()),
			},
		},
	)
	s.Require().NoError(err, "client.GetMonthEvents must not have error")
	s.Require().Equal(ev.EventID, events[0].EventID, "client.GetMonthEvents proper event")
}

func (s *IntegrationSuite) Test_CreateEventWithNotify() {
	ownerID := uuid.NewString()

	now := time.Now()

	startAt := time.Now().UTC().Add(100*24*time.Hour + 10*time.Second)
	ev := &v1.Event{
		EventID:      uuid.NewString(),
		StartAt:      timestamppb.New(startAt),
		EndAt:        timestamppb.New(startAt.Add(12 * time.Hour)),
		Title:        "event with notify",
		NotifyBefore: 100,
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout)
	defer cancel()

	_, err := s.Client.CreateEvent(ctx, ownerID, ev)
	s.Require().NoError(err, "event with notify must be created")

	if d := time.Since(now); d > 10*time.Second {
		s.Failf("creating takes too much time", "takes %s", d.String())
	}

	<-time.After(10 * time.Second)

	fmt.Printf("send notification for ownerID=%s eventID=%s:\n", ownerID, ev.EventID)
}
