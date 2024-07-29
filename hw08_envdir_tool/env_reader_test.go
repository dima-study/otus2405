package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type tableTest struct {
	name   string
	dir    string
	result Environment
	err    error
	errIs  error
	errAs  func(err error) (error, bool)
}

func genErrAs[T error](target T) func(err error) (error, bool) {
	return func(err error) (error, bool) {
		return target, errors.As(err, &target)
	}
}

func TestReadDir(t *testing.T) {
	tests := []tableTest{
		{
			name: "t1",
			dir:  "./testdata/t1",
			result: Environment{
				"FOO":       EnvValue{"BAR", false},
				"MULTILINE": EnvValue{"Line 1", false},
				"UNSET":     EnvValue{"", true},
			},
			err:   nil,
			errAs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ReadDir(tt.dir)

			if tt.err != nil || tt.errAs != nil {
				if err == nil {
					t.Fatal("error must not be empty, but it is")
				}

				if tt.err != nil && !errors.Is(err, tt.err) {
					t.Errorf("error must be %#v, but got %#v", tt.err, err)
				}

				if tt.errAs != nil {
					if target, ok := tt.errAs(err); !ok {
						t.Errorf("error must be %#v, but got %#v", target, err)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("error must be empty, but got %#v", err)
				}
			}

			if tt.result != nil && result == nil {
				t.Fatal("result must not be empty, but it is")
			}

			if tt.result == nil && result != nil {
				t.Fatalf("result must be empty, but got %#v", result)
			}

			require.Equal(t, tt.result, result, "proper result")
		})
	}
}
