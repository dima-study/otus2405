package hw09structvalidator

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_structValidator_Supports(t *testing.T) {
	type Struct struct{}

	var (
		St Struct
		st struct{}
		i  int
		s  string
		ii []int
		ss []string
	)

	v := StructValidator()

	tests := []struct {
		name      string
		field     any
		supported bool
	}{
		{"Struct", St, true},
		{"struct", st, true},
		{"int", i, false},
		{"string", s, false},
		{"[]string", ii, false},
		{"[]int", ss, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testType := reflect.TypeOf(tt.field)

			if tt.supported {
				require.True(t, v.Supports(testType), "must be supported")
			} else {
				require.False(t, v.Supports(testType), " must not be supported")
			}
		})
	}
}

func Test_structValidator_ValidatorsFor(t *testing.T) {
	v := StructValidator(
		StringValidator(),
		IntValidator(),
		SliceValidator(
			StringValidator(),
			IntValidator(),
		),
	)

	t.Run("success", func(t *testing.T) {
		s := struct {
			S  string   `validate:"regexp:^\\d+$"`
			Ss []string `validate:"len:3"`
			I  int      `validate:"min:3"`
			Ii []int    `validate:"max:3"`
		}{}

		sT := reflect.TypeOf(s)
		validators, err := v.ValidatorsFor(sT, []Rule{{RuleStructNested, ""}})

		require.Nil(t, err, "err must be nil")
		require.Len(t, validators, 1, "must be 1 validator")
	})

	t.Run("success not supported, but not exported", func(t *testing.T) {
		s := struct {
			s  string   `validate:"reGexp:^\\d+$"`
			ss []string `validate:"lEn:3"`
			i  int      `validate:"mIn:3"`
			ii []int    `validate:"mAx:3"`
		}{}

		sT := reflect.TypeOf(s)
		validators, err := v.ValidatorsFor(sT, []Rule{{RuleStructNested, ""}})

		require.Nil(t, err, "err must be nil")
		require.Len(t, validators, 1, "must be 1 validator")
	})

	t.Run("failed not supported", func(t *testing.T) {
		s := struct {
			S  string   `validate:"reGexp:^\\d+$"`
			Ss []string `validate:"lEn:3"`
			I  int      `validate:"mIn:3"`
			Ii []int    `validate:"mAx:3"`
		}{}

		sT := reflect.TypeOf(s)
		validators, err := v.ValidatorsFor(sT, []Rule{{RuleStructNested, ""}})

		require.ErrorIs(t, err, ErrStructNested, "err is ErrStructNested")
		require.ErrorIs(t, err, ErrRuleNotSupported, "err is ErrValidatorRuleNotSupported")
		require.Len(t, validators, 0, "must be 0 validators")
	})

	t.Run("failed rule not supported", func(t *testing.T) {
		s := struct {
			S  string   `validate:"regexp=^\\d+$"`
			Ss []string `validate:"len=3"`
			I  int      `validate:"min=3"`
			Ii []int    `validate:"max=3"`
		}{}

		sT := reflect.TypeOf(s)
		validators, err := v.ValidatorsFor(sT, []Rule{{RuleStructNested, ""}})

		t.Log(err)

		require.ErrorIs(t, err, ErrStructNested, "err is ErrStructNested")
		require.ErrorIs(t, err, ErrRuleNotSupported, "err is ErrRuleNotSupported")
		require.Len(t, validators, 0, "must be 0 validators")
	})

	t.Run("failed incorrect syntax", func(t *testing.T) {
		s := struct {
			B bool `validate:""`
		}{}

		sT := reflect.TypeOf(s)
		validators, err := v.ValidatorsFor(sT, []Rule{{RuleStructNested, ""}})

		require.ErrorIs(t, err, ErrStructNested, "err is ErrStructNested")
		require.ErrorIs(t, err, ErrTypeNotSupported, "err is ErrTypeNotSupported")
		require.Len(t, validators, 0, "must be 0 validators")
	})
}

func Test_structValidator_validatorNested(t *testing.T) {
	SV := StructValidator(
		StringValidator(),
		IntValidator(),
		SliceValidator(
			StringValidator(),
			IntValidator(),
		),
	)

	v, ok := SV.(structValidator)
	require.True(t, ok, "must be instance of structValidator")

	type SubStruct struct {
		S  string   `validate:"regexp:^\\d+$"`
		Ss []string `validate:"len:3"`
		I  int      `validate:"min:3"`
		Ii []int    `validate:"max:3"`
	}

	type Struct struct {
		S  string    `validate:"regexp:^\\d+$"`
		Ss []string  `validate:"len:3"`
		I  int       `validate:"min:3"`
		Ii []int     `validate:"max:3"`
		St SubStruct `valdate:"nested"`
	}
	s := Struct{}

	sT := reflect.TypeOf(s)

	fieldValidators, err := v.ValidatorsFor(sT, []Rule{{RuleStructNested, ""}})
	t.Log(err)
	require.Nil(t, err, "err must be nil")
	require.Len(t, fieldValidators, 1, "must be 1 validator")

	t.Run("validate", func(t *testing.T) {
		fn := fieldValidators[0]

		tests := []struct {
			val     Struct
			err     bool
			numErr  int
			errItem []int
		}{
			{
				val: Struct{
					S:  "1234",
					Ss: []string{"aaa", "bbb", "ccc"},
					I:  3,
					Ii: []int{1, 2, 3},
					St: SubStruct{
						S: "4321",
						I: 5,
					},
				},
				err:    false,
				numErr: 0,
			},
			{
				val: Struct{
					S:  "abcd",
					Ss: []string{"a", "bbb", "c"},
					I:  1,
					Ii: []int{1, 22, 333, 4444},
				},
				err:     true,
				numErr:  4,
				errItem: []int{},
			},
		}

		for _, tt := range tests {
			t.Run(fmt.Sprintf("%v", tt.val), func(t *testing.T) {
				val := reflect.ValueOf(tt.val)
				err = fn(val)
				if tt.err {
					t.Log(err)

					var asErr StructErrors
					require.ErrorAs(t, err, &asErr, "must be StructErrors")
					require.Len(t, asErr, tt.numErr, "must be correct number of errors")
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

func Test_structValidator_getFieldsValidators(t *testing.T) {
	SV := StructValidator(
		StringValidator(),
		IntValidator(),
		SliceValidator(
			StringValidator(),
			IntValidator(),
		),
	)

	v, ok := SV.(structValidator)
	require.True(t, ok, "must be instance of structValidator")

	type Struct struct {
		B  bool
		S  string `validate:"regexp:^\\d+$"`
		Ns string
		Ss []string `validate:"len:3"`
		I  int      `validate:"min:3|max:5"`
		Ii []int    `validate:"max:3"`
	}
	s := Struct{}

	sT := reflect.TypeOf(s)

	fieldValidators, err := v.getFieldsValidators(sT)
	require.Nil(t, err, "err must be nil")
	require.Len(t, fieldValidators, 4, "must be 1 validator")
}
