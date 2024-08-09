package hw09structvalidator

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_stringValidator_Supports(t *testing.T) {
	s := struct {
		a any
		i int
		s string
	}{}

	sT := reflect.TypeOf(s)
	v := StringValidator()

	tests := []struct {
		field     string
		supported bool
	}{
		{"a", false},
		{"i", false},
		{"s", true},
	}

	for _, tt := range tests {
		t.Run("field "+tt.field, func(t *testing.T) {
			field, ok := sT.FieldByName(tt.field)
			require.True(t, ok, "field "+tt.field+" must be found")

			if tt.supported {
				require.True(t, v.Supports(field.Type), "field "+tt.field+" must be supported")
			} else {
				require.False(t, v.Supports(field.Type), "field "+tt.field+" must not be supported")
			}
		})
	}
}

func Test_stringValidator_ValidatorsFor(t *testing.T) {
	s := struct {
		field string
	}{}

	sT := reflect.TypeOf(s)
	v := StringValidator()

	structField, ok := sT.FieldByName("field")
	require.True(t, ok, "field must be found")

	t.Run("success", func(t *testing.T) {
		validators, err := v.ValidatorsFor(
			structField.Type,
			[]Rule{
				{RuleStringLen, "2"},
				{RuleStringIn, "aa,bb,cc"},
				{RuleStringRegexp, "^[^0-9]{2}$"},
			},
		)
		require.Nil(t, err, "err must be nil")
		require.Len(t, validators, 3, "must be 3 validators")
	})

	t.Run("failed not supported", func(t *testing.T) {
		validators, err := v.ValidatorsFor(
			structField.Type,
			[]Rule{
				{"lEn", "2"},
				{"iN", "aa,bb,cc"},
				{"reGexp", "^[^0-9]{2}$"},
			},
		)
		require.ErrorIs(t, err, ErrRuleNotSupported, "err is ErrValidatorRuleNotSupported")
		require.Len(t, validators, 0, "must be 0 validators")
	})

	t.Run("failed invalid rule condition", func(t *testing.T) {
		validators, err := v.ValidatorsFor(
			structField.Type,
			[]Rule{
				{RuleStringLen, "abc"},
			},
		)
		t.Log(err)
		require.ErrorIs(t, err, ErrRuleInvalidCondition, "err is ErrRuleInvalidCondition")
		require.Len(t, validators, 0, "must be 0 validators")
	})
}

func Test_stringValidator_validatorLen(t *testing.T) {
	v := stringValidator{}
	t.Run("error", func(t *testing.T) {
		fn, err := v.validatorLen("invalid")
		require.NotNil(t, err, "err must not be nil")
		require.Nil(t, fn, "fn must be nil")
	})

	t.Run("validate", func(t *testing.T) {
		fn, err := v.validatorLen("2")
		require.Nil(t, err, "err must be nil")
		require.NotNil(t, fn, "fn must not be nil")

		tests := []struct {
			val string
			err bool
		}{
			{"a", true},
			{"aa", false},
			{"aaa", true},
		}

		for _, tt := range tests {
			t.Run(tt.val, func(t *testing.T) {
				val := reflect.ValueOf(tt.val)
				err = fn(val)
				if tt.err {
					require.ErrorIs(t, err, ErrStringLen, "err must be ErrStringLen")
				} else {
					require.Nil(t, err, "err must be nil")
				}
			})
		}
	})
}

func Test_stringValidator_validatorRegexp(t *testing.T) {
	v := stringValidator{}
	t.Run("error", func(t *testing.T) {
		fn, err := v.validatorRegexp("[a-b")
		require.NotNil(t, err, "err must not be nil")
		require.Nil(t, fn, "fn must be nil")
	})

	t.Run("validate", func(t *testing.T) {
		fn, err := v.validatorRegexp(`^[^0-9]{2}$`)
		require.Nil(t, err, "err must be nil")
		require.NotNil(t, fn, "fn must not be nil")

		tests := []struct {
			val string
			err bool
		}{
			{"a", true},
			{"1", true},
			{"aa", false},
			{"11", true},
			{"aaa", true},
			{"111", true},
		}

		for _, tt := range tests {
			t.Run(tt.val, func(t *testing.T) {
				val := reflect.ValueOf(tt.val)
				err = fn(val)
				if tt.err {
					require.ErrorIs(t, err, ErrStringRegexp, "err must be ErrStringRegexp")
				} else {
					require.Nil(t, err, "err must be nil")
				}
			})
		}
	})
}

func Test_stringValidator_validatorIn(t *testing.T) {
	v := stringValidator{}

	t.Run("validate", func(t *testing.T) {
		fn, err := v.validatorIn("bb,ccc,a")
		require.Nil(t, err, "err must be nil")
		require.NotNil(t, fn, "fn must not be nil")

		tests := []struct {
			val string
			err bool
		}{
			{"a", false},
			{"aa", true},
			{"c", true},
			{"ccc", false},
			{"cccc", true},
			{"b", true},
			{"bb", false},
			{"bbb", true},
		}

		for _, tt := range tests {
			t.Run(tt.val, func(t *testing.T) {
				val := reflect.ValueOf(tt.val)
				err = fn(val)
				if tt.err {
					require.ErrorIs(t, err, ErrStringIn, "err must be ErrStringIn")
				} else {
					require.Nil(t, err, "err must be nil")
				}
			})
		}
	})
}
