package hw02unpackstring

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnpack(t *testing.T) {
	type subtest struct {
		input    string
		expected string
	}
	tests := []struct {
		name     string
		subtests []subtest
	}{
		{
			name: "general",
			subtests: []subtest{
				{input: "a4bc2d5e", expected: "aaaabccddddde"},
				{input: "abccd", expected: "abccd"},
				{input: "", expected: ""},
				{input: "a", expected: "a"},
				{input: "aaa0b", expected: "aab"},
				{input: "-1", expected: "-"},
				{input: "d\n5abc", expected: "d\n\n\n\n\nabc"},
				{input: "aaa0bc0", expected: "aab"},
				{input: "世0☺1-1☺1", expected: "☺-☺"},
				{input: "\x003", expected: "\x00\x00\x00"},
			},
		},
		{
			name: "with escaping",
			subtests: []subtest{
				{input: `qwe\4\5`, expected: `qwe45`},
				{input: `qwe\45`, expected: `qwe44444`},
				{input: `qwe\\5`, expected: `qwe\\\\\`},
				{input: `qwe\\\3`, expected: `qwe\3`},
				{input: `\\\\`, expected: `\\`},
				{input: `\\1`, expected: `\`},
				{input: `\\0`, expected: ""},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for _, stc := range tc.subtests {
				t.Run(stc.input, func(t *testing.T) {
					result, err := Unpack(stc.input)
					require.NoError(t, err)
					require.Equal(t, stc.expected, result)
				})
			}
		})
	}
}

func TestUnpackInvalidString(t *testing.T) {
	invalid := []struct {
		s   string
		err error
	}{
		{"3abc", ErrNothingToRepeat},
		{"45", ErrNothingToRepeat},
		{"aaa10b", ErrNothingToRepeat},
		{"00", ErrNothingToRepeat},
		{"1a", ErrNothingToRepeat},
		{"1", ErrNothingToRepeat},

		{`\`, ErrNothingToEscape},

		{`\a`, ErrInvalidEscapeSymbol},
	}
	for _, tc := range invalid {
		tc := tc
		t.Run(tc.s, func(t *testing.T) {
			_, err := Unpack(tc.s)
			require.Truef(t, errors.Is(err, tc.err), "actual error %q", err)
		})
	}
}

func Test_digit(t *testing.T) {
	tests := []struct {
		name string
		c    rune
		want int
	}{
		{"0=0", '0', 0},
		{"4=4", '4', 4},
		{"9=9", '9', 9},
		{"a=-1", 'a', -1},
		{"\\0=-1", 0, -1},
		{"\\n=-1", '\n', -1},
		{"\\\\=-1", '\\', -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := digit(tt.c)
			require.Equal(t, tt.want, got)
		})
	}
}

func BenchmarkRepeatStrings(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := rune(b.N % math.MaxInt32)
		bldr := strings.Builder{}
		bldr.Grow(10)
		bldr.WriteString(strings.Repeat(string(b), 10))
		_ = bldr.String()
	}
}

func BenchmarkRepeatLoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := rune(b.N % math.MaxInt32)
		bldr := strings.Builder{}
		bldr.Grow(10)
		for range 10 {
			bldr.WriteRune(b)
		}
		_ = bldr.String()
	}
}

func BenchmarkRepeatLoopMake(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := rune(b.N % math.MaxInt32)
		s := make([]rune, 10)
		for i := range 10 {
			s[i] = b
		}
		bldr := strings.Builder{}
		bldr.Grow(10)
		bldr.WriteString(string(s))
		_ = bldr.String()
	}
}
