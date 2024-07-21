package main

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

// Copy tries to copy up to limit bytes from fromPath (with offset) into toPath.
// Returns any occurred error.
func Copy(fromPath, toPath string, offset, limit int64) error {
	// Quick-check if file (not dir!) exists and offset is not exceeds file size.
	fInfo, err := os.Stat(fromPath)
	switch {
	case err != nil:
		return fmt.Errorf("can't get file info: %w", err)
	case fInfo.IsDir():
		return fmt.Errorf("%w: '%s' is a directory", ErrUnsupportedFile, fromPath)
	case fInfo.Size() < offset:
		return fmt.Errorf(
			"%w: filesize=%d < offset=%d",
			ErrOffsetExceedsFileSize, fInfo.Size(), offset,
		)
	}

	return nil
}
