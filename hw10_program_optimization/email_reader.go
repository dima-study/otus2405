package hw10programoptimization

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/goccy/go-json"
)

// EmailReader tries to read one email per line from the reader r.
// Line must be a valid json string.
type EmailReader struct {
	sc *bufio.Scanner
}

func NewEmailReader(r io.Reader) *EmailReader {
	sc := bufio.NewScanner(r)
	return &EmailReader{sc: sc}
}

var ErrEmpty = errors.New("empty read")

// readLine tries to read the next line from the reader.
//
// Returns:
//   - read line on success
//   - nil on EOF
//   - occurred error
//   - ErrEmpty if read line is empty
func (r *EmailReader) readLine() ([]byte, error) {
	// Try to read the line.
	if !r.sc.Scan() {
		// Return error if occurred.
		if err := r.sc.Err(); err != nil {
			return nil, err
		}

		// Return nil on EOF
		return nil, nil
	}

	line := r.sc.Bytes()

	// Line is empty
	if len(line) == 0 {
		return nil, ErrEmpty
	}

	return line, nil
}

// NextEmail tries to extract and return email from the next json-line.
// Returns email (possibly empty) in lower case on success.
// Returns empty string and io.EOF on EOF, or any occurred error.
func (r *EmailReader) NextEmail() (string, error) {
	// UserEmail is being used to extract email.
	// Not exported.
	type UserEmail struct {
		Email string
	}

	line, err := r.readLine()
	if err != nil {
		return "", fmt.Errorf("read line: %w", err)
	}

	// It is EOF
	if line == nil {
		return "", io.EOF
	}

	// Extract email from the line
	var userEmail UserEmail

	err = json.Unmarshal(line, &userEmail)
	if err != nil {
		return "", fmt.Errorf("unmarshal: %w", err)
	}

	return strings.ToLower(userEmail.Email), nil
}
