package hw02unpackstring

import (
	"errors"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

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
	var sym rune
	for _, sym = range str {
		if nextToEscape {
			escaped = true
			nextToEscape = false
		}

		// Symbol is ASCII: possible escape or digit.
		if sym < 256 {
			if sym == '\\' {
				nextToEscape = true
			} else {
				dgt = digit(byte(sym))
			}
		}

		// Current symbol sym should be escaped.
		if escaped {
			if dgt != -1 {
				// It is escaped digit.
				symToUnpack = sym
				dgt = -1
			} else if sym == '\\' {
				// It is escaped backslash.
				symToUnpack = sym
			} else {
				// Only digit or backslash(\) could be escaped.
				return "", ErrInvalidString
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
			return "", ErrInvalidString
		}

		// Okay, unpack symbol to output
		for range dgt {
			output.WriteRune(symToUnpack)
		}

		// Reset reads.
		symToUnpack = startReadState
		dgt = -1
	}

	// Edge-case: current non-written symbol is the last symbol in str and it is not a digit:
	//   - write it to output when next should not be escaped;
	//   - otherwise return error.
	if symToUnpack != startReadState {
		if nextToEscape {
			return "", ErrInvalidString
		}
		output.WriteRune(sym)
	}

	return output.String(), nil
}

// digit returns int digit (0-9) from provided  ASCII char c.
// Returns -1 if it is not a digit.
func digit(c byte) int {
	if '0' <= c && c <= '9' {
		return int(c - '0')
	}

	return -1
}
