package hw02unpackstring

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidEscapeSymbol = errors.New("invalid symbol to escape")
	ErrNothingToRepeat     = errors.New("nothing to repeat")
	ErrNothingToEscape     = errors.New("nothing to escape")
)

func Unpack(str string) (string, error) {
	output := strings.Builder{}

	// startReadState indicates there are no symbol read yet,
	// and next read is the starting one.
	const startReadState rune = -1

	// escaped indicates that symToUnpack should be escaped.
	escaped := false

	// nextToEscape indicates that next symbol should be escaped:
	// when true, escaped flag will be set to true on next symbol read.
	nextToEscape := false

	symToUnpack := startReadState // symToUnpack holds symbol to be unpacked (symbol at previous position).
	dgt := int(-1)                // -1 means current symbol is not a digit.

	// For each symbol sym in str.
	var i int
	var sym rune
	for i, sym = range str {
		if nextToEscape {
			escaped = true
			nextToEscape = false
		}

		// Check if escape or digit.
		if sym == '\\' {
			nextToEscape = true
		} else {
			dgt = digit(sym)
		}

		// Current symbol sym should be escaped.
		if escaped {
			switch {
			case dgt != -1:
				// It is escaped digit.
				symToUnpack = sym
				dgt = -1
			case sym == '\\':
				// It is escaped backslash.
				symToUnpack = sym
			default:
				// Only digit or backslash(\) could be escaped.
				return "", fmt.Errorf(
					"%w: symbol %q at position %d",
					ErrInvalidEscapeSymbol, sym, i,
				)
			}

			// Now escaped and ready-to-unpack symbol is in symToUnpack.
			// Reset escape flags and do next read.
			escaped = false
			nextToEscape = false
			continue
		}

		if dgt == -1 {
			// CASE: current symbol is not a digit (so no need to repeat previous one symbol).
			// So if symToUnpack is not in startReadState, write it to output.
			if symToUnpack != startReadState {
				output.WriteRune(symToUnpack)
			}

			// Then make current symbol to be next symbol to unpack.
			symToUnpack = sym

			continue
		}

		// CASE: current symbol is a digit (so need to repeat previous one symbol).

		// If symToUnpack is in startReadState: digit has been read, but no previous symbol read.
		// Return error.
		if symToUnpack == startReadState {
			return "", fmt.Errorf("%w: at position %d", ErrNothingToRepeat, i)
		}

		// Okay, unpack symbol to output
		output.WriteString(strings.Repeat(string(symToUnpack), dgt))

		// Reset reads.
		symToUnpack = startReadState
		dgt = -1
	}

	// Edge-case: current non-written symbol is the last symbol in str and it is not a digit:
	//   - write it to output when next should not be escaped;
	//   - otherwise return error.
	if symToUnpack != startReadState {
		if nextToEscape {
			return "", fmt.Errorf("%w: at position %d", ErrNothingToEscape, i)
		}
		output.WriteRune(sym)
	}

	return output.String(), nil
}

// digit returns int digit (0-9) from provided rune r.
// Returns -1 if it is not a digit.
func digit(r rune) int {
	if '0' <= r && r <= '9' {
		return int(r - '0')
	}

	return -1
}
