package pg

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	model "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/event"
)

// TEST_STORAGE_PG env var to run the test.
// Example value: "postgresql://calendar:calendar@localhost:5432/calendar"
const envVarName = "TEST_STORAGE_PG"

var dataSource = os.Getenv(envVarName)

type populateArgs struct {
	now      time.Time
	times    [][2]time.Time
	eventIDs [3]model.ID
	ownerIDs [3]model.OwnerID
}

func mkEvent(
	t *testing.T,
	eventID model.ID,
	ownerID model.OwnerID,
	title model.Title,
	startAt time.Time,
	endAt time.Time,
) model.Event {
	t.Helper()

	event, err := model.NewEvent(eventID, ownerID, title, startAt, endAt)
	require.NoErrorf(t, err, "must not have error while create an event %s", title)
	return event
}

type PgTestSuite struct {
	suite.Suite

	storage *Storage
	args    populateArgs
}

func (s *PgTestSuite) SetupTest() {
	storage, err := NewStorage(dataSource)
	s.Require().NoError(err, "must connect")
	s.storage = storage
	s.populate()
}

func (s *PgTestSuite) TearDownTest() {
	s.storage.DB.MustExec("TRUNCATE events")
	s.storage.DB.Close()
	s.storage = nil
}

func TestPgSuite(t *testing.T) {
	if dataSource == "" {
		t.Skip("set env var " + envVarName + " to the data-source string to run the test!")
	}

	suite.Run(t, new(PgTestSuite))
}

func (s *PgTestSuite) populate() {
	/*
	*    ...... 1h ...... ... 1h ... ...... 1h ...... ... 1h ... ...... 1h ......
	*    [now+1h, now+2h) .......... [now+3h, now+4h) .......... [now+5h, now+6h)
	 */
	now := time.Now()
	times := [][2]time.Time{
		{now.Add(1 * time.Hour), now.Add(2 * time.Hour)},
		{now.Add(3 * time.Hour), now.Add(4 * time.Hour)},
		{now.Add(5 * time.Hour), now.Add(6 * time.Hour)},
	}
	eventIDs := [...]model.ID{model.NewID(), model.NewID(), model.NewID()}
	ownerIDs := [...]model.OwnerID{model.NewOwnerID(), model.NewOwnerID(), model.NewOwnerID()}

	tests := []struct {
		name  string
		event model.Event
		err   error
	}{
		{
			name:  "populate event#1 user#1",
			event: mkEvent(s.T(), eventIDs[0], ownerIDs[0], "1", times[1][0], times[1][1]),
			err:   nil,
		},
		{
			name:  "populate event#1 user#1 duplicate",
			event: mkEvent(s.T(), eventIDs[0], ownerIDs[0], "1", times[1][0], times[1][1]),
			err:   model.ErrEventAlreadyExists,
		},
		{
			name:  "populate event#2 user#1 time overlap",
			event: mkEvent(s.T(), eventIDs[1], ownerIDs[0], "2", times[1][0], times[1][1]),
			err:   model.ErrTimeIsBusy,
		},
		{
			name:  "populate event#2 user#1",
			event: mkEvent(s.T(), eventIDs[1], ownerIDs[0], "2", times[0][0], times[0][1]),
			err:   nil,
		},
		{
			name:  "populate event#1 user#3",
			event: mkEvent(s.T(), eventIDs[0], ownerIDs[2], "1", times[1][0], times[1][1]),
			err:   nil,
		},
		{
			name:  "populate event#2 user#3",
			event: mkEvent(s.T(), eventIDs[1], ownerIDs[2], "2", times[0][0], times[0][1]),
			err:   nil,
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			err := s.storage.AddEvent(context.Background(), tt.event)
			if tt.err == nil {
				require.NoError(t, err, "must not have error")
			} else {
				require.Error(t, err, "must have error")
				require.ErrorIsf(t, err, tt.err, "error must be %v", tt.err)
			}
		})
	}

	s.args = populateArgs{
		now:      now,
		times:    times,
		eventIDs: eventIDs,
		ownerIDs: ownerIDs,
	}
}

func (s *PgTestSuite) Test_FindEvent() {
	type args struct {
		ownerID model.OwnerID
		eventID model.ID
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "user#1 event#1",
			args: args{s.args.ownerIDs[0], s.args.eventIDs[0]},
			err:  nil,
		},
		{
			name: "user#1 event#2",
			args: args{s.args.ownerIDs[0], s.args.eventIDs[1]},
			err:  nil,
		},
		{
			name: "user#1 event#3",
			args: args{s.args.ownerIDs[0], s.args.eventIDs[2]},
			err:  model.ErrEventNotFound,
		},
		{
			name: "user#2 event#1",
			args: args{s.args.ownerIDs[1], s.args.eventIDs[0]},
			err:  model.ErrEventNotFound,
		},
		{
			name: "user#3 event#1",
			args: args{s.args.ownerIDs[2], s.args.eventIDs[0]},
			err:  nil,
		},
		{
			name: "user#3 event#2",
			args: args{s.args.ownerIDs[2], s.args.eventIDs[1]},
			err:  nil,
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			event, err := s.storage.FindEvent(
				context.Background(),
				tt.args.ownerID,
				tt.args.eventID,
			)

			if tt.err == nil {
				require.NoError(t, err, "must not have error")
				require.EqualValues(t, tt.args.ownerID, event.OwnerID(), "ownerID must be equal")
				require.EqualValues(t, tt.args.eventID, event.EventID(), "eventID must be equal")
			} else {
				require.Error(t, err, "must have error")
				require.ErrorIsf(t, err, tt.err, "error must be %v", tt.err)
			}
		})
	}
}

func (s *PgTestSuite) Test_UpdateEvent() {
	s.T().Run("success", func(t *testing.T) {
		t.Run("order before update", func(t *testing.T) {
			events, err := s.storage.QueryEvents(
				context.Background(),
				s.args.ownerIDs[0],
				time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
				time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC),
			)
			require.NoError(s.T(), err, "must hot have error")
			require.Len(t, events, 2, "must have 2 events for user #1")
			require.Equal(t, events[0].EventID(), s.args.eventIDs[1], "event #2 must be at 0")
			require.Equal(t, events[1].EventID(), s.args.eventIDs[0], "event #1 must be at 1")
		})

		event := mkEvent(t, s.args.eventIDs[1], s.args.ownerIDs[0], "2", s.args.times[2][0], s.args.times[2][1])

		err := s.storage.UpdateEvent(context.Background(), event)
		require.NoError(t, err, "must not have error")
	})

	s.T().Run("not found", func(t *testing.T) {
		event := mkEvent(t, s.args.eventIDs[2], s.args.ownerIDs[0], "2", s.args.times[2][0], s.args.times[2][1])

		err := s.storage.UpdateEvent(context.Background(), event)
		require.Error(t, err, "must have error")
		require.ErrorIs(t, err, model.ErrEventNotFound, "must be ErrEventNotFound error")
	})

	s.T().Run("time is busy", func(t *testing.T) {
		event := mkEvent(t, s.args.eventIDs[1], s.args.ownerIDs[0], "2", s.args.times[1][0], s.args.times[1][1])

		err := s.storage.UpdateEvent(context.Background(), event)
		require.Error(t, err, "must have error")
		require.ErrorIs(t, err, model.ErrTimeIsBusy, "must be ErrTimeIsBusy error")
	})

	s.T().Run("order after update", func(t *testing.T) {
		events, err := s.storage.QueryEvents(
			context.Background(),
			s.args.ownerIDs[0],
			time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC),
		)
		require.NoError(s.T(), err, "must hot have error")
		require.Len(t, events, 2, "must have 2 events for user #1")
		require.Equal(t, events[0].EventID(), s.args.eventIDs[0], "event #1 must be at 0")
		require.Equal(t, events[1].EventID(), s.args.eventIDs[1], "event #2 must be at 1")
	})
}

func (s *PgTestSuite) Test_DeleteEvent() {
	s.T().Run("no event for unknown user", func(t *testing.T) {
		err := s.storage.DeleteEvent(context.Background(), model.NewOwnerID(), model.NewID())
		require.Error(t, err, "must have error")
		require.ErrorIs(t, err, model.ErrEventNotFound, "must be ErrEventNotFound error")
	})

	s.T().Run("no event for user", func(t *testing.T) {
		err := s.storage.DeleteEvent(context.Background(), s.args.ownerIDs[0], model.NewID())
		require.Error(t, err, "must have error")
		require.ErrorIs(t, err, model.ErrEventNotFound, "must be ErrEventNotFound error")
	})

	s.T().Run("success", func(t *testing.T) {
		err := s.storage.DeleteEvent(context.Background(), s.args.ownerIDs[0], s.args.eventIDs[0])
		require.NoError(t, err, "must not have error")

		events, err := s.storage.QueryEvents(
			context.Background(),
			s.args.ownerIDs[0],
			time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
			time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC),
		)
		require.NoError(s.T(), err, "must hot have error")
		require.Len(t, events, 1, "must have 1 event after delete")
		require.Equal(t, events[0].EventID(), s.args.eventIDs[1], "event #2 must be at 0")
	})
}

func (s *PgTestSuite) Test_QueryEvents() {
	/*
	         [=============]
	  [---]                   [---]    before-after
	  [------]             [------]    before edge-after edge
	  [----------]     [----------]    before overlap startAt-after overlap endAt
	         [---]     [---]           startAt inside-inside entAt
	         [-------------]           equal
	         [--------------------]    startAt overlap endAt
	  [--------------------]           overlap startAt endAt
	*/

	tests := []struct {
		name     string
		from     time.Time
		to       time.Time
		evendIDs []model.ID
	}{
		{
			name:     "before 1",
			from:     s.args.times[0][0].Add(-time.Minute),
			to:       s.args.times[0][0].Add(-time.Second),
			evendIDs: []model.ID{},
		},
		{
			name:     "after 1",
			from:     s.args.times[0][1].Add(time.Second),
			to:       s.args.times[0][1].Add(time.Minute),
			evendIDs: []model.ID{},
		},
		{
			name:     "before edge 1",
			from:     s.args.times[0][0].Add(-time.Second),
			to:       s.args.times[0][0],
			evendIDs: []model.ID{},
		},
		{
			name:     "after edge 1",
			from:     s.args.times[0][1],
			to:       s.args.times[0][1].Add(time.Second),
			evendIDs: []model.ID{},
		},
		{
			name:     "before overlap startAt 1",
			from:     s.args.times[0][0].Add(-time.Second),
			to:       s.args.times[0][0].Add(time.Second),
			evendIDs: []model.ID{s.args.eventIDs[1]},
		},
		{
			name:     "after overlap endAt 1",
			from:     s.args.times[0][1].Add(-time.Second),
			to:       s.args.times[0][1].Add(time.Second),
			evendIDs: []model.ID{s.args.eventIDs[1]},
		},
		{
			name:     "startAt inside 1",
			from:     s.args.times[0][0],
			to:       s.args.times[0][0].Add(time.Second),
			evendIDs: []model.ID{s.args.eventIDs[1]},
		},
		{
			name:     "inside endAt 1",
			from:     s.args.times[0][1].Add(-time.Second),
			to:       s.args.times[0][1],
			evendIDs: []model.ID{s.args.eventIDs[1]},
		},
		{
			name:     "equal 1",
			from:     s.args.times[0][0],
			to:       s.args.times[0][1],
			evendIDs: []model.ID{s.args.eventIDs[1]},
		},
		{
			name:     "startAt overlap endAt 1",
			from:     s.args.times[0][0],
			to:       s.args.times[0][1].Add(time.Second),
			evendIDs: []model.ID{s.args.eventIDs[1]},
		},
		{
			name:     "overlap startAt endAt 1",
			from:     s.args.times[0][0].Add(-time.Second),
			to:       s.args.times[0][1],
			evendIDs: []model.ID{s.args.eventIDs[1]},
		},
		{
			name:     "first+second overlap",
			from:     s.args.times[0][0].Add(-time.Second),
			to:       s.args.times[1][0].Add(time.Second),
			evendIDs: []model.ID{s.args.eventIDs[1], s.args.eventIDs[0]},
		},
		{
			name:     "first+second edge",
			from:     s.args.times[0][1],
			to:       s.args.times[1][0],
			evendIDs: []model.ID{},
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			events, err := s.storage.QueryEvents(context.Background(), s.args.ownerIDs[0], tt.from, tt.to)
			eventIDs := []model.ID{}

			require.NoError(s.T(), err, "must not have error")

			for _, e := range events {
				eventIDs = append(eventIDs, e.EventID())
			}

			require.Equal(t, tt.evendIDs, eventIDs, "proper result")
		})
	}
}
