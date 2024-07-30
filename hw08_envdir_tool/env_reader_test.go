package main

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
)

type tableTest struct {
	name   string
	dir    string
	result Environment
	err    error
}

func TestReadDir(t *testing.T) {
	tests := []tableTest{
		{
			name: "t",
			dir:  "./testdata/t",
			result: Environment{
				"EMPTY":         EnvValue{"", false},
				"FOO":           EnvValue{"BAR", false},
				"MULTILINE":     EnvValue{"Line 1", false},
				"SPACES":        EnvValue{"  Spaces", false},
				"SPACES-N-TABS": EnvValue{"  \tspaces and tabs", false},
				"TABS":          EnvValue{"\tTabs", false},
				"TABS-N-SPACES": EnvValue{"\t  tabs and spaces", false},
				"UNSET":         EnvValue{"", true},
				"ZEROES":        EnvValue{"  here\nnew\nline", false},
			},
			err: nil,
		},
		{
			name:   "env dir doesn't exists",
			dir:    "./testdata/does-not-exist",
			result: nil,
			err:    fs.ErrNotExist,
		},
		{
			name:   "empty env dir",
			dir:    "./testdata/empty",
			result: Environment{},
			err:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ReadDir(tt.dir)

			if tt.err != nil {
				if err == nil {
					t.Fatal("error must not be empty, but it is")
				}

				if tt.err != nil && !errors.Is(err, tt.err) {
					t.Errorf("error must be %#v, but got %#v", tt.err, err)
				}
			} else if err != nil {
				t.Fatalf("error must be empty, but got %#v", err)
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
