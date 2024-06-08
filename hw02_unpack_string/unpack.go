package hw02unpackstring

import (
	"errors"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(_ string) (string, error) {
	// Place your code here.
	return "", nil
}

// digit returns int digit (0-9) from provided  ASCII char c.
// Returns -1 if it is not a digit.
func digit(c byte) int {
	if '0' <= c && c <= '9' {
		return int(c - '0')
	}

	return -1
}
