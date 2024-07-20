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

func Copy(fromPath, toPath string, offset, limit int64) error {
	fInfo, err := os.Stat(fromPath)
	if err != nil {
		return fmt.Errorf("can't get file info: %w", err)
	}

	if fInfo.IsDir() {
		return fmt.Errorf("%w: '%s' is a directory", ErrUnsupportedFile, fromPath)
	}

	if fInfo.Size() < offset {
		return fmt.Errorf(
			"%w: filesize=%d < offset=%d",
			ErrOffsetExceedsFileSize, fInfo.Size(), offset,
		)
	}

	return nil
}
