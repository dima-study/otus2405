package event

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func mkEventTitle(t *testing.T, title string) Title {
	t.Helper()

	eventTitle, err := NewTitle(title)
	require.NoError(t, err, "must not have error")

	return eventTitle
}

func TestNewEvent(t *testing.T) {
	type args struct {
		eventID ID
		ownerID OwnerID
		title   Title
		startAt time.Time
		endAt   time.Time
	}

	now := time.Now()

	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			args: args{
				eventID: NewID(),
				ownerID: NewOwnerID(),
				title:   mkEventTitle(t, "ok"),
				startAt: now,
				endAt:   now.Add(time.Second),
			},
			err: nil,
		},
		{
			name: "startAt = endAt",
			args: args{
				eventID: NewID(),
				ownerID: NewOwnerID(),
				title:   mkEventTitle(t, "ok"),
				startAt: now,
				endAt:   now,
			},
			err: ErrTimeEndBeforeStart,
		},
		{
			name: "endAt before startAt",
			args: args{
				eventID: NewID(),
				ownerID: NewOwnerID(),
				title:   mkEventTitle(t, "ok"),
				startAt: now.Add(time.Second),
				endAt:   now,
			},
			err: ErrTimeEndBeforeStart,
		},
	}
	for _, tt := range tests {
		testName := tt.name
		if testName == "" {
			testName = string(tt.args.title)
		}
		t.Run(testName, func(t *testing.T) {
			got, err := NewEvent(tt.args.eventID, tt.args.ownerID, tt.args.title, tt.args.startAt, tt.args.endAt)

			if tt.err == nil {
				require.NoError(t, err, "must not have error")

				require.Equal(t, tt.args.eventID, got.EventID(), "EventID must equal to ID")
				require.Equal(t, tt.args.ownerID, got.OwnerID(), "UserID must equal to ownerID")

				require.True(t, got.StartAt().Equal(tt.args.startAt), "StartAt must equal to startAt")
				require.True(t, got.EndAt().Equal(tt.args.endAt), "EndAt must equal to endAt")
			} else {
				require.Error(t, err, "must have error")
				require.ErrorIsf(t, err, tt.err, "must be %v", tt.err)
			}
		})
	}
}

func TestNewEventIDFromString(t *testing.T) {
	_, err := NewIDFromString(uuid.NewString())
	require.NoError(t, err, "must not have error")

	_, err = NewIDFromString("invalid")
	require.Error(t, err, "must have error")
	require.ErrorIs(t, err, ErrInvalidEventID, "must be ErrInvalidEventID error")
}

func TestNewOwnerIDFromString(t *testing.T) {
	_, err := NewOwnerIDFromString(uuid.NewString())
	require.NoError(t, err, "must not have error")

	_, err = NewOwnerIDFromString("invalid")
	require.Error(t, err, "must have error")
	require.ErrorIs(t, err, ErrInvalidOwnerID, "must be ErrInvalidOwnerID error")
}

func TestNewEventTitle(t *testing.T) {
	_, err := NewTitle(strings.Repeat("x", MaxEventTitleLen))
	require.NoError(t, err, "must not have error")

	_, err = NewTitle(strings.Repeat("x", MaxEventTitleLen+1))
	require.Error(t, err, "must have error")
	require.ErrorIs(t, err, ErrMaxTitleLen, "must be ErrMaxTitleLen error")

	_, err = NewTitle("")
	require.Error(t, err, "must have error")
	require.ErrorIs(t, err, ErrEmptyTitle, "must be ErrEmptyTitle error")
}
