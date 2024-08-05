package hw09structvalidator

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_intValidator_Supports(t *testing.T) {
	s := struct {
		a any
		i int
		s string
	}{}

	sT := reflect.TypeOf(s)
	v := IntValidator()

	tests := []struct {
		field     string
		supported bool
	}{
		{"a", false},
		{"i", true},
		{"s", false},
	}

	for _, tt := range tests {
		t.Run("field "+tt.field, func(t *testing.T) {
			field, ok := sT.FieldByName(tt.field)
			require.True(t, ok, "field "+tt.field+" must be found")

			if tt.supported {
				require.True(t, v.Supports(field), "field "+tt.field+" must be supported")
			} else {
				require.False(t, v.Supports(field), "field "+tt.field+" must not be supported")
			}
		})
	}
}

func Test_intValidator_ValidatorsFor(t *testing.T) {
	v := IntValidator()

	t.Run("success", func(t *testing.T) {
		validators, err := v.ValidatorsFor(`max:42|in:42,24,33|min:24`)
		require.Nil(t, err, "err must be nil")
		require.Len(t, validators, 3, "must be 3 validators")
	})

	t.Run("failed not supported", func(t *testing.T) {
		validators, err := v.ValidatorsFor(`maX:42|In:42,24,33|mIn:24`)
		require.ErrorIs(t, err, ErrValidatorRuleNotSupported, "err is ErrValidatorRuleNotSupported")
		require.Len(t, validators, 0, "must be 0 validators")
	})

	t.Run("failed incorrect syntax", func(t *testing.T) {
		validators, err := v.ValidatorsFor(`maX=42`)
		require.ErrorIs(t, err, ErrValidatorIncorrectRuleSyntax, "err is ErrValidatorIncorrectRuleSyntax")
		require.Len(t, validators, 0, "must be 0 validators")
	})
}

func Test_intValidator_validatorMin(t *testing.T) {
	v := intValidator{}
	t.Run("error", func(t *testing.T) {
		fn, err := v.validatorMin("invalid")
		require.NotNil(t, err, "err must not be nil")
		require.Nil(t, fn, "fn must be nil")
	})

	t.Run("validate", func(t *testing.T) {
		fn, err := v.validatorMin("42")
		require.Nil(t, err, "err must be nil")
		require.NotNil(t, fn, "fn must not be nil")

		tests := []struct {
			val int
			err error
		}{
			{5, ErrIntMin},
			{42, nil},
			{45, nil},
		}

		for _, tt := range tests {
			t.Run(strconv.Itoa(tt.val), func(t *testing.T) {
				val := reflect.ValueOf(tt.val)
				err = fn(val)
				if tt.err != nil {
					require.ErrorIs(t, err, tt.err, "err must be ErrIntMin")
				} else {
					require.Nil(t, err, "err must be nil")
				}
			})
		}
	})
}

func Test_intValidator_validatorMax(t *testing.T) {
	v := intValidator{}
	t.Run("error", func(t *testing.T) {
		fn, err := v.validatorMax("invalid")
		require.NotNil(t, err, "err must not be nil")
		require.Nil(t, fn, "fn must be nil")
	})

	t.Run("validate", func(t *testing.T) {
		fn, err := v.validatorMax("42")
		require.Nil(t, err, "err must be nil")
		require.NotNil(t, fn, "fn must not be nil")

		tests := []struct {
			val int
			err error
		}{
			{5, nil},
			{42, nil},
			{45, ErrIntMax},
		}

		for _, tt := range tests {
			t.Run(strconv.Itoa(tt.val), func(t *testing.T) {
				val := reflect.ValueOf(tt.val)
				err = fn(val)
				if tt.err != nil {
					require.ErrorIs(t, err, tt.err, "err must be ErrIntMin")
				} else {
					require.Nil(t, err, "err must be nil")
				}
			})
		}
	})
}

func Test_intValidator_validatorIn(t *testing.T) {
	v := intValidator{}
	t.Run("error", func(t *testing.T) {
		fn, err := v.validatorIn("invalid")
		require.NotNil(t, err, "err must not be nil")
		require.Nil(t, fn, "fn must be nil")
	})

	t.Run("validate", func(t *testing.T) {
		fn, err := v.validatorIn("42,24")
		require.Nil(t, err, "err must be nil")
		require.NotNil(t, fn, "fn must not be nil")

		tests := []struct {
			val int
			err error
		}{
			{5, ErrIntIn},
			{24, nil},
			{42, nil},
		}

		for _, tt := range tests {
			t.Run(strconv.Itoa(tt.val), func(t *testing.T) {
				val := reflect.ValueOf(tt.val)
				err = fn(val)
				if tt.err != nil {
					require.ErrorIs(t, err, tt.err, "err must be ErrIntMin")
				} else {
					require.Nil(t, err, "err must be nil")
				}
			})
		}
	})
}
