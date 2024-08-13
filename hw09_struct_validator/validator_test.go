package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `          validate:"min:18|max:50"`
		Email  string          `          validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `          validate:"in:admin,stuff"`
		Phones []string        `          validate:"len:11|regexp:^\\d+$"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `                          json:"omitempty"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		in      interface{}
		errIs   error
		testErr func(t *testing.T, err error)
	}{
		{
			in:    UserRole("test"),
			errIs: ErrTypeNotSupported,
		},
		{
			in: User{
				ID:     "ca5b933b-88a6-4769-9f60-44624c01859a",
				Name:   "Name",
				Age:    25,
				Email:  "user@example.com",
				Role:   "stuff",
				Phones: []string{"12345678901"},
			},
		},
		{
			in: User{
				ID:     "invalid",
				Age:    16,
				Email:  "netu pochty :-(",
				Role:   "superadmin",
				Phones: []string{"12345678901", "sebe zvoni!"},
			},
			testErr: func(t *testing.T, err error) {
				t.Helper()

				var verr ValidationErrors

				require.ErrorAs(t, err, &verr, "must be ValidationErrors")
				require.Len(t, verr, 5, "must be 5 underlying error")

				errList := []struct {
					field string
					errIs error
				}{
					{"ID", ErrStringLen},
					{"Age", ErrIntMin},
					{"Email", ErrStringRegexp},
					{"Role", ErrStringIn},
				}
				for i, fieldErr := range errList {
					require.Equal(t, verr[i].Field, fieldErr.field, "must be field "+fieldErr.field)
					require.ErrorIs(t, verr[i].Err, fieldErr.errIs, "must be correct error")
				}

				require.Equal(t, verr[4].Field, "Phones", "must be field Phones")

				var sliceErr SliceErrors
				require.ErrorAs(t, verr[4].Err, &sliceErr, "must be SliceErrors")
				require.Len(t, sliceErr, 1, "must be 1 slice error")
				require.Equal(t, 1, sliceErr[0].Index, "must be error at pos 1")
				require.ErrorIs(t, sliceErr[0].Err, ErrStringRegexp, "must be ErrStringRegexp")
			},
		},
		{
			in: App{
				Version: "1.120",
			},
		},
		{
			in: Response{
				Code: 404,
				Body: "not found",
			},
		},
		{
			in: Response{
				Code: 301,
				Body: "redirect",
			},
			testErr: func(t *testing.T, err error) {
				t.Helper()

				var verr ValidationErrors

				require.ErrorAs(t, err, &verr, "must be ValidationErrors")
				require.Len(t, verr, 1, "must be 1 underlying error")

				require.Equal(t, verr[0].Field, "Code", "must be field Code")
				require.ErrorIs(t, verr[0].Err, ErrIntIn, "must be ErrIntIn")
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)

			switch {
			case (tt.errIs != nil || tt.testErr != nil) && err == nil:
				t.Error("wants error, but there are no")
			case tt.errIs == nil && tt.testErr == nil && err != nil:
				t.Errorf("doesn't want error, but got %q", err)
			default:
				switch {
				case tt.errIs != nil && !errors.Is(err, tt.errIs):
					t.Errorf("wants error %#v, but got %#v", tt.errIs, err)
				case tt.testErr != nil:
					tt.testErr(t, err)
				}
			}
		})
	}
}
