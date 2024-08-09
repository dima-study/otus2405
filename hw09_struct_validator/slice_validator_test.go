package hw09structvalidator

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_sliceValidator_Supports(t *testing.T) {
	s := struct {
		a any

		i int
		s string

		ii []int
		ss []string
	}{}

	sT := reflect.TypeOf(s)
	v := SliceValidator(IntValidator(), StringValidator())

	tests := []struct {
		field     string
		supported bool
	}{
		{"a", false},
		{"i", false},
		{"s", false},
		{"ii", true},
		{"ss", true},
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

func Test_sliceValidator_ValidatorsFor(t *testing.T) {
	s := struct {
		field []string
	}{}

	sT := reflect.TypeOf(s)
	v := SliceValidator(StringValidator())

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
		require.Len(t, validators, 1, "must be 1 validator")
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
		require.ErrorIs(t, err, ErrRuleInvalidCondition, "err is ErrRuleInvalidCondition")
		require.Len(t, validators, 0, "must be 0 validators")
	})
}

func Test_sliceValidator_validatorSlice(t *testing.T) {
	s := struct {
		field []string
	}{}

	sT := reflect.TypeOf(s)
	SV := SliceValidator(StringValidator())
	v, ok := SV.(sliceValidator)
	require.True(t, ok, "must be instance of sliceValidator")

	structField, ok := sT.FieldByName("field")
	require.True(t, ok, "field must be found")

	fieldValidators, err := v.ValidatorsFor(
		structField.Type,
		[]Rule{
			{RuleStringLen, "2"},
			{RuleStringIn, "a,1,aa,bb,cc,12,ddd"},
			{RuleStringRegexp, "^[^0-9]+$"},
		},
	)
	require.Nil(t, err, "err must be nil")
	require.Len(t, fieldValidators, 1, "must be 1 validator")

	t.Run("validate", func(t *testing.T) {
		fn := fieldValidators[0]

		tests := []struct {
			val     []string
			err     bool
			numErr  int
			errItem []int
		}{
			{[]string{"aa"}, false, 0, nil},
			{[]string{"aa", "12"}, true, 1, []int{0, 1}},
			{[]string{"dd", "aa"}, true, 1, []int{1, 0}},
			{[]string{"d1", "aa", "a"}, true, 3, []int{1, 0, 1}},
			{[]string{"1", "aa", "b", "3"}, true, 7, []int{1, 0, 1, 1}},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("%v", tt.val), func(t *testing.T) {
				val := reflect.ValueOf(tt.val)
				err = fn(val)
				if tt.err {
					var asErr SliceErrors
					require.ErrorAs(t, err, &asErr, "must be SliceErrors")
					require.Len(t, asErr, tt.numErr, "must be correct number of errors")

					for _, e := range asErr {
						require.Equal(t, 1, tt.errItem[e.Index], "item error must have correct index")
					}

					t.Log(asErr)
				} else {
					if err != nil {
						t.Log(err)
					}
					require.Nil(t, err, "err must be nil")
				}
			})
		}
	})
}
