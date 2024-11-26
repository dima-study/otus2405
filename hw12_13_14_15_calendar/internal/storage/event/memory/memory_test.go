package memory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	model "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/model/event"
	modelStorage "github.com/dima-study/otus2405/hw12_13_14_15_calendar/internal/storage/event"
)

type populateArgs struct {
	now      time.Time
	times    [][2]time.Time
	eventIDs [3]model.ID
	ownerIDs [3]model.OwnerID
}

func mkTitle(t *testing.T, title string) model.Title {
	t.Helper()

	eventTitle, err := model.NewTitle(title)
	require.NoErrorf(t, err, "must not have error while create an event title %s", title)
	return eventTitle
}

func mkEvent(
	t *testing.T,
	eventID model.ID,
	ownerID model.OwnerID,
	title model.Title,
	startAt time.Time,
	endAt time.Time,
	notifyBefore uint,
) model.Event {
	t.Helper()

	event, err := model.NewEvent(eventID, ownerID, title, startAt, endAt)
	require.NoErrorf(t, err, "must not have error while create an event %s", title)

	event.NotifyBefore = notifyBefore

	return event
}

func populate(t *testing.T) (*Storage, populateArgs) {
	t.Helper()

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
			name:  "event#1 user#1",
			event: mkEvent(t, eventIDs[0], ownerIDs[0], "1", times[1][0], times[1][1], 5),
			err:   nil,
		},
		{
			name:  "event#1 user#1 duplicate",
			event: mkEvent(t, eventIDs[0], ownerIDs[0], "1", times[1][0], times[1][1], 5),
			err:   modelStorage.ErrEventAlreadyExists,
		},
		{
			name:  "event#2 user#1 time overlap",
			event: mkEvent(t, eventIDs[1], ownerIDs[0], "2", times[1][0], times[1][1], 5),
			err:   modelStorage.ErrTimeIsBusy,
		},
		{
			name:  "event#2 user#1",
			event: mkEvent(t, eventIDs[1], ownerIDs[0], "2", times[0][0], times[0][1], 0),
			err:   nil,
		},
		{
			name:  "event#1 user#3",
			event: mkEvent(t, eventIDs[0], ownerIDs[2], "1", times[1][0], times[1][1], 5),
			err:   nil,
		},
		{
			name:  "event#2 user#3",
			event: mkEvent(t, eventIDs[1], ownerIDs[2], "2", times[0][0], times[0][1], 0),
			err:   nil,
		},
	}

	storage := NewStorage()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.AddEvent(context.Background(), tt.event)
			if tt.err == nil {
				require.NoError(t, err, "must not have error")
			} else {
				require.Error(t, err, "must have error")
				require.ErrorIsf(t, err, tt.err, "error must be %v", tt.err)
			}
		})
	}

	return storage, populateArgs{
		now:      now,
		times:    times,
		eventIDs: eventIDs,
		ownerIDs: ownerIDs,
	}
}

func TestMemory_AddEvent(t *testing.T) {
	storage, pargs := populate(t)

	events := storage.userMap[pargs.ownerIDs[0]]
	require.Len(t, events, 2, "must have 2 events for user #1")

	t.Run("proper order", func(t *testing.T) {
		require.Equal(t, 0, findEventIndex(events, pargs.eventIDs[1]), "event #2 must be at 0")
		require.Equal(t, 1, findEventIndex(events, pargs.eventIDs[0]), "event #1 must be at 1")
	})

	events = storage.userMap[pargs.ownerIDs[2]]
	require.Len(t, events, 2, "must have 2 events for user #3")
}

func TestMemory_FindEvent(t *testing.T) {
	storage, pargs := populate(t)

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
			args: args{pargs.ownerIDs[0], pargs.eventIDs[0]},
			err:  nil,
		},
		{
			name: "user#1 event#2",
			args: args{pargs.ownerIDs[0], pargs.eventIDs[1]},
			err:  nil,
		},
		{
			name: "user#1 event#3",
			args: args{pargs.ownerIDs[0], pargs.eventIDs[2]},
			err:  modelStorage.ErrEventNotFound,
		},
		{
			name: "user#2 event#1",
			args: args{pargs.ownerIDs[1], pargs.eventIDs[0]},
			err:  modelStorage.ErrEventNotFound,
		},
		{
			name: "user#3 event#1",
			args: args{pargs.ownerIDs[2], pargs.eventIDs[0]},
			err:  nil,
		},
		{
			name: "user#3 event#2",
			args: args{pargs.ownerIDs[2], pargs.eventIDs[1]},
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := storage.FindEvent(
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

func TestMemory_UpdateEvent(t *testing.T) {
	storage, pargs := populate(t)

	t.Run("success", func(t *testing.T) {
		events := storage.userMap[pargs.ownerIDs[0]]
		require.Len(t, events, 2, "must have 2 events for user #1")

		event := mkEvent(t, pargs.eventIDs[1], pargs.ownerIDs[0], "2", pargs.times[2][0], pargs.times[2][1], 0)

		err := storage.UpdateEvent(context.Background(), event)
		require.NoError(t, err, "must not have error")

		t.Run("proper order", func(t *testing.T) {
			require.Equal(t, 0, findEventIndex(events, pargs.eventIDs[0]), "event #1 must be at 0")
			require.Equal(t, 1, findEventIndex(events, pargs.eventIDs[1]), "event #2 must be at 1")
		})
	})

	t.Run("not found", func(t *testing.T) {
		event := mkEvent(t, pargs.eventIDs[2], pargs.ownerIDs[0], "2", pargs.times[2][0], pargs.times[2][1], 0)

		err := storage.UpdateEvent(context.Background(), event)
		require.Error(t, err, "must have error")
		require.ErrorIs(t, err, modelStorage.ErrEventNotFound, "must be ErrEventNotFound error")
	})

	t.Run("time is busy", func(t *testing.T) {
		event := mkEvent(t, pargs.eventIDs[1], pargs.ownerIDs[0], "2", pargs.times[1][0], pargs.times[1][1], 0)

		err := storage.UpdateEvent(context.Background(), event)
		require.Error(t, err, "must have error")
		require.ErrorIs(t, err, modelStorage.ErrTimeIsBusy, "must be ErrTimeIsBusy error")
	})

	events := storage.userMap[pargs.ownerIDs[0]]
	t.Run("order not changed", func(t *testing.T) {
		require.Equal(t, 0, findEventIndex(events, pargs.eventIDs[0]), "event #1 must be at 0")
		require.Equal(t, 1, findEventIndex(events, pargs.eventIDs[1]), "event #2 must be at 1")
	})
}

func TestMemory_DeleteEvent(t *testing.T) {
	storage, pargs := populate(t)

	t.Run("no event for unknown user", func(t *testing.T) {
		err := storage.DeleteEvent(context.Background(), model.NewOwnerID(), model.NewID())
		require.Error(t, err, "must have error")
		require.ErrorIs(t, err, modelStorage.ErrEventNotFound, "must be ErrEventNotFound error")
	})

	t.Run("no event for user", func(t *testing.T) {
		err := storage.DeleteEvent(context.Background(), pargs.ownerIDs[0], model.NewID())
		require.Error(t, err, "must have error")
		require.ErrorIs(t, err, modelStorage.ErrEventNotFound, "must be ErrEventNotFound error")
	})

	t.Run("success", func(t *testing.T) {
		err := storage.DeleteEvent(context.Background(), pargs.ownerIDs[0], pargs.eventIDs[0])
		require.NoError(t, err, "must not have error")

		events := storage.userMap[pargs.ownerIDs[0]]
		require.Len(t, events, 1, "proper len")
		require.Equal(t, 0, findEventIndex(events, pargs.eventIDs[1]), "event #2 must be at 0")
	})
}

func TestMemory_QueryEvents(t *testing.T) {
	storage, pargs := populate(t)

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
			from:     pargs.times[0][0].Add(-time.Minute),
			to:       pargs.times[0][0].Add(-time.Second),
			evendIDs: []model.ID{},
		},
		{
			name:     "after 1",
			from:     pargs.times[0][1].Add(time.Second),
			to:       pargs.times[0][1].Add(time.Minute),
			evendIDs: []model.ID{},
		},
		{
			name:     "before edge 1",
			from:     pargs.times[0][0].Add(-time.Second),
			to:       pargs.times[0][0],
			evendIDs: []model.ID{},
		},
		{
			name:     "after edge 1",
			from:     pargs.times[0][1],
			to:       pargs.times[0][1].Add(time.Second),
			evendIDs: []model.ID{},
		},
		{
			name:     "before overlap startAt 1",
			from:     pargs.times[0][0].Add(-time.Second),
			to:       pargs.times[0][0].Add(time.Second),
			evendIDs: []model.ID{pargs.eventIDs[1]},
		},
		{
			name:     "after overlap endAt 1",
			from:     pargs.times[0][1].Add(-time.Second),
			to:       pargs.times[0][1].Add(time.Second),
			evendIDs: []model.ID{pargs.eventIDs[1]},
		},
		{
			name:     "startAt inside 1",
			from:     pargs.times[0][0],
			to:       pargs.times[0][0].Add(time.Second),
			evendIDs: []model.ID{pargs.eventIDs[1]},
		},
		{
			name:     "inside endAt 1",
			from:     pargs.times[0][1].Add(-time.Second),
			to:       pargs.times[0][1],
			evendIDs: []model.ID{pargs.eventIDs[1]},
		},
		{
			name:     "equal 1",
			from:     pargs.times[0][0],
			to:       pargs.times[0][1],
			evendIDs: []model.ID{pargs.eventIDs[1]},
		},
		{
			name:     "startAt overlap endAt 1",
			from:     pargs.times[0][0],
			to:       pargs.times[0][1].Add(time.Second),
			evendIDs: []model.ID{pargs.eventIDs[1]},
		},
		{
			name:     "overlap startAt endAt 1",
			from:     pargs.times[0][0].Add(-time.Second),
			to:       pargs.times[0][1],
			evendIDs: []model.ID{pargs.eventIDs[1]},
		},
		{
			name:     "first+second overlap",
			from:     pargs.times[0][0].Add(-time.Second),
			to:       pargs.times[1][0].Add(time.Second),
			evendIDs: []model.ID{pargs.eventIDs[1], pargs.eventIDs[0]},
		},
		{
			name:     "first+second edge",
			from:     pargs.times[0][1],
			to:       pargs.times[1][0],
			evendIDs: []model.ID{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, _ := storage.QueryEvents(context.Background(), pargs.ownerIDs[0], tt.from, tt.to)
			eventIDs := []model.ID{}

			for _, e := range events {
				eventIDs = append(eventIDs, e.EventID())
			}

			require.Equal(t, tt.evendIDs, eventIDs, "proper result")
		})
	}
}

func Test_findNewEventIndex(t *testing.T) {
	storage, pargs := populate(t)
	events := storage.userMap[pargs.ownerIDs[0]]

	require.Len(t, events, 2, "proper len")

	tests := []struct {
		name    string
		startAt time.Time
		endAt   time.Time
		pos     int
	}{
		{
			name:    "add at pos 0",
			startAt: pargs.times[0][0].Add(-time.Second),
			endAt:   pargs.times[0][0],
			pos:     0,
		},
		{
			name:    "add at pos 1",
			startAt: pargs.times[0][1],
			endAt:   pargs.times[1][0],
			pos:     1,
		},
		{
			name:    "add at pos 2",
			startAt: pargs.times[1][1],
			endAt:   pargs.times[1][1].Add(time.Second),
			pos:     2,
		},
		{
			name:    "can't add at start",
			startAt: pargs.times[0][0].Add(-time.Second),
			endAt:   pargs.times[0][0].Add(time.Second),
			pos:     -1,
		},
		{
			name:    "can't add in-between",
			startAt: pargs.times[0][1].Add(-time.Second),
			endAt:   pargs.times[1][0],
			pos:     -1,
		},
		{
			name:    "add at end",
			startAt: pargs.times[1][1].Add(-time.Second),
			endAt:   pargs.times[1][1].Add(time.Second),
			pos:     -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := findNewEventIndex(
				events,
				mkEvent(
					t,
					model.NewID(),
					model.NewOwnerID(),
					mkTitle(t, tt.name),
					tt.startAt,
					tt.endAt,
					0,
				),
			)
			require.Equal(t, tt.pos, i, "proper value")
		})
	}
}

func Test_PurgeOldEvents(t *testing.T) {
	storage, pargs := populate(t)

	err := storage.PurgeOldEvents(
		context.Background(),
		pargs.times[1][0].Truncate(time.Second),
	)
	require.NoError(t, err, "must not have arror")

	count := 0
	for _, userEvents := range storage.userMap {
		count += len(userEvents)
	}
	require.Equal(t, 2, count, "must be proper value")
}

func TestMemory_QueryEventsToNotify(t *testing.T) {
	storage, pargs := populate(t)

	tests := []struct {
		name     string
		from     time.Time
		to       time.Time
		evendIDs []model.ID
	}{
		{
			name: "need notify",
			from: pargs.times[1][0].Truncate(time.Second).Add(-5 * 24 * time.Hour),
			to:   pargs.times[1][0].Truncate(time.Second).Add(-5*24*time.Hour + time.Hour),
			evendIDs: []model.ID{
				pargs.eventIDs[0],
				pargs.eventIDs[0],
			},
		},
		// {
		// 	name:     "no need notify",
		// 	from:     pargs.times[0][0].Truncate(time.Second).Add(-5 * 24 * time.Hour),
		// 	to:       pargs.times[0][0].Truncate(time.Second).Add(-5*24*time.Hour + time.Hour),
		// 	evendIDs: []model.ID{},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, err := storage.QueryEventsToNotify(context.Background(), tt.from, tt.to)
			require.NoError(t, err, "must not have arror")

			eventIDs := []model.ID{}
			for _, e := range events {
				eventIDs = append(eventIDs, e.EventID())
			}

			require.Equal(t, tt.evendIDs, eventIDs, "proper result")
		})
	}
}
