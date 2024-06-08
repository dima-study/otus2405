package hw02unpackstring

import (
	"errors"
	"strings"
	"unicode/utf8"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(str string) (string, error) {
	output := strings.Builder{}

	// startReadState indicates there are no symbol read yet,
	// and next read is the starting one.
	const startReadState rune = -1

	symToUnpack := startReadState // symToUnpack holds symbol to be unpacked.
	dgt := int(-1)                // -1 means current symbol is not a digit.

	// Store str length to check the last symbol.
	strLen := len(str)

	// For each symbol in str.
	for i, sym := range str {
		// symbol is ASCII: possible digit
		if sym < 256 {
			dgt = digit(byte(sym))
		}

		if dgt == -1 {
			// Current symbol is not a digit (so no need to repeat previous one symbol).
			// So if symToUnpack is not in startReadState, write it to output.
			// Then make current symbol to be next symbol to unpack.
			if symToUnpack != startReadState {
				output.WriteRune(symToUnpack)
			}
			symToUnpack = sym

			// Edge-case: current symbol is the last symbol in str: write it to output.
			symLen := utf8.RuneLen(sym)
			if i+symLen == strLen {
				output.WriteRune(sym)
			}

			continue
		}

		// Current symbol is a digit (so need to repeat previous one symbol).

		// If symToUnpack is in startReadState, return error.
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
