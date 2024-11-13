package calendar

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

	proto "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/api/proto/event/v1"
	calendarBusiness "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/business/calendar"
	"github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/grpc/auth"
	model "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/event"
	memoryStorage "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/storage/event/memory"
	pgStorage "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/storage/event/pg"
)

const envVarName = "TEST_STORAGE_PG"

var dataSource = os.Getenv(envVarName)

func Test_APISuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}

type APITestSuite struct {
	suite.Suite

	app     *App
	storage model.Storage

	ownerID  model.OwnerID
	eventIDs []model.ID

	mx sync.Mutex
}

func (s *APITestSuite) SetupSuite() {
	var storage model.Storage

	if dataSource != "" {
		var err error
		storage, err = pgStorage.NewStorage(dataSource)
		s.Require().NoError(err, "must connect")
	} else {
		storage = memoryStorage.NewStorage()
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	business := calendarBusiness.NewApp(logger, storage)
	app := NewApp(business, logger)

	s.app = app
	s.storage = storage
	s.ownerID = model.NewOwnerID()
}

func (s *APITestSuite) CreateEvent(startAt time.Time, endAt time.Time) (*proto.Event, model.Event) {
	eventStr := uuid.NewString()

	ctx, err := auth.WithOwnerID(context.Background(), string(s.ownerID))
	s.Require().NoError(err, "auth.WithOwnerID must not have error")

	ownerID, err := auth.OwnerIDFromContext(ctx)
	s.Require().NoError(err, "auth.OwnerIDFromContext must not have error")
	s.Require().Equal(string(s.ownerID), string(ownerID), "ownerID must be equal")

	protoEvent := proto.Event{
		EventID: eventStr,
		StartAt: &timestamppb.Timestamp{
			Seconds: startAt.Unix(),
			Nanos:   int32(startAt.Nanosecond()),
		},
		EndAt: &timestamppb.Timestamp{
			Seconds: endAt.Unix(),
			Nanos:   int32(endAt.Nanosecond()),
		},
		Title:        "event title",
		Description:  "",
		NotifyBefore: 0,
	}
	req := &proto.CreateEventRequest{Event: &protoEvent}

	_, err = s.app.CreateEvent(ctx, req)
	s.Require().NoError(err, "app.CreateEvent must not have error")

	event, err := s.storage.FindEvent(ctx, ownerID, model.ID(eventStr))
	s.Require().NoError(err, "must not have error")

	s.Require().True(event.StartAt().Equal(startAt), "event.StartAt must equal")
	s.Require().True(event.EndAt().Equal(endAt), "event.EndAt must equal")

	{
		s.mx.Lock()
		s.eventIDs = append(s.eventIDs, event.EventID())
		s.mx.Unlock()
	}

	return &protoEvent, event
}

func (s *APITestSuite) Test_CreateEvent() {
	_, event := s.CreateEvent(time.Now(), time.Now().Add(2*time.Hour))
	require.NotEmpty(s.T(), event, "must not be empty")
}

func (s *APITestSuite) Test_UpdateEvent() {
	var protoEvent *proto.Event
	var event model.Event
	s.Run("create event", func() {
		when := time.Now().Add(time.Hour * 48)
		protoEvent, event = s.CreateEvent(when, when.Add(2*time.Hour))
	})

	s.Require().NotNil(protoEvent, "protoEvent must not be nil")

	ctx, err := auth.WithOwnerID(context.Background(), string(s.ownerID))
	s.Require().NoError(err, "auth.WithOwnerID must not have error")

	protoEvent.Title = "changed title"
	req := &proto.UpdateEventRequest{Event: protoEvent}
	_, err = s.app.UpdateEvent(ctx, req)
	s.Require().NoError(err, "app.UpdateEvent must not have error")

	storageEvent, err := s.storage.FindEvent(ctx, s.ownerID, event.EventID())
	s.Require().NoError(err, "must not have error")

	event.Title = model.Title("changed title")
	s.Require().Equal(event, storageEvent, "event is changed")
}

func (s *APITestSuite) Test_DeleteEvent() {
	var protoEvent *proto.Event
	s.Run("create event", func() {
		when := time.Now().Add(time.Hour * 96)
		protoEvent, _ = s.CreateEvent(when, when.Add(2*time.Hour))
	})
	s.Require().NotNil(protoEvent, "protoEvent must not be nil")

	ctx, err := auth.WithOwnerID(context.Background(), string(s.ownerID))
	s.Require().NoError(err, "auth.WithOwnerID must not have error")

	{
		s.mx.Lock()
		for _, eventID := range s.eventIDs {
			req := &proto.DeleteEventRequest{EventID: string(eventID)}

			_, err := s.app.DeleteEvent(ctx, req)
			s.Require().NoError(err, "app.DeleteEvent must not have error")

			_, err = s.storage.FindEvent(ctx, s.ownerID, eventID)
			s.Require().Error(err, "must have error")
			s.Require().ErrorIs(err, model.ErrEventNotFound, "must have model.ErrEventNotFound error")
		}

		for _, eventID := range s.eventIDs {
			req := &proto.DeleteEventRequest{EventID: string(eventID)}

			_, err := s.app.DeleteEvent(ctx, req)
			s.Require().Error(err, "app.DeleteEvent must have error after second delete")
		}

		s.eventIDs = []model.ID{}
		s.mx.Unlock()
	}
}

func (s *APITestSuite) Test_GetDayEvents() {
	var events []model.Event
	s.Run("create 60 events", func() {
		now := time.Now()
		for i := range 60 {
			startAt := time.Date(now.Year()+1, now.Month(), now.Day()+i, 10, 10, 0, 0, time.UTC)
			_, event := s.CreateEvent(startAt, startAt.Add(2*time.Hour))
			events = append(events, event)
		}
	})
	s.Require().Len(events, 60, "must be created 60 events")

	ctx, err := auth.WithOwnerID(context.Background(), string(s.ownerID))
	s.Require().NoError(err, "auth.WithOwnerID must not have error")

	for _, event := range events {
		req := &proto.GetDayEventsRequest{
			Day: &proto.Date{
				Year:  int32(event.StartAt().Year()),
				Month: int32(event.StartAt().Month()),
				Day:   int32(event.StartAt().Day()),
			},
		}

		resp, err := s.app.GetDayEvents(ctx, req)
		s.Require().NoError(err, "app.GetDayEvents must not have error")
		s.Require().Len(resp.Events, 1, "must be 1 event in day")

		for _, e := range resp.Events {
			respEvent, err := protoToModel(e, s.ownerID)
			s.Require().NoError(err, "protoToModel must not have error")

			s.Require().Equal(event, respEvent, "event from response must be equal")
		}
	}
}

func (s *APITestSuite) Test_GetWeekEvents() {
	var events []model.Event
	s.Run("create 60 events", func() {
		now := time.Now()
		for i := range 60 {
			startAt := time.Date(now.Year()+2, now.Month(), now.Day()+i, 10, 10, 0, 0, time.UTC)
			_, event := s.CreateEvent(startAt, startAt.Add(2*time.Hour))
			events = append(events, event)
		}
	})
	s.Require().Len(events, 60, "must be created 60 events")

	ctx, err := auth.WithOwnerID(context.Background(), string(s.ownerID))
	s.Require().NoError(err, "auth.WithOwnerID must not have error")

	for i, event := range events {
		req := &proto.GetWeekEventsRequest{
			StartDay: &proto.Date{
				Year:  int32(event.StartAt().Year()),
				Month: int32(event.StartAt().Month()),
				Day:   int32(event.StartAt().Day()),
			},
		}

		resp, err := s.app.GetWeekEvents(ctx, req)
		s.Require().NoError(err, "app.GetWeekEvents must not have error")

		n := min(len(events)-i, 7)
		s.Require().Lenf(resp.Events, n, "must be %d events in week", n)

		for j, re := range resp.Events {
			respEvent, err := protoToModel(re, s.ownerID)
			s.Require().NoError(err, "protoToModel must not have error")

			s.Require().Equal(events[i+j], respEvent, "event from response must be equal")
		}
	}
}

func (s *APITestSuite) Test_GetMonthEvents() {
	eventsInMonth := map[time.Month]int{}
	var events []model.Event
	s.Run("create 60 events", func() {
		now := time.Now()
		for i := range 60 {
			startAt := time.Date(now.Year()+3, now.Month(), now.Day()+i, 10, 10, 0, 0, time.UTC)
			_, event := s.CreateEvent(startAt, startAt.Add(2*time.Hour))
			events = append(events, event)

			eventsInMonth[startAt.Month()]++
		}
	})
	s.Require().Len(events, 60, "must be created 60 events")

	ctx, err := auth.WithOwnerID(context.Background(), string(s.ownerID))
	s.Require().NoError(err, "auth.WithOwnerID must not have error")

	for _, event := range events {
		req := &proto.GetMonthEventsRequest{
			Month: &proto.Month{
				Year:  int32(event.StartAt().Year()),
				Month: int32(event.StartAt().Month()),
			},
		}

		resp, err := s.app.GetMonthEvents(ctx, req)
		s.Require().NoError(err, "app.GetMonthEvents must not have error")

		s.Require().Lenf(
			resp.Events,
			eventsInMonth[event.StartAt().Month()],
			"must be %d events in month",
			eventsInMonth[event.StartAt().Month()],
		)
	}
}
