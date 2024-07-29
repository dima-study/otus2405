package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	pb "github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

// Copy tries to copy up to limit bytes from fromPath (with offset) into toPath.
// Returns any occurred error.
//
// Expects non-negative offset and limit.
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

	// Open fromPath for reading...
	fileFrom, err := os.OpenFile(fromPath, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("can't open input file for reading: %w", err)
	}
	defer fileFrom.Close() // Should we catch the error?

	// ... and seeks the offset.
	_, err = fileFrom.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("can't seek offset %d of input file: %w", offset, err)
	}

	// Open toPath for writing (create if not exists, truncate if exists).
	fileTo, err := os.OpenFile(toPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("can't open output file for writing: %w", err)
	}
	defer fileTo.Close() // Should we catch the error?

	// Limit up to fromPath file size.
	if limit == 0 {
		limit = fInfo.Size()
	}

	// fileFromLimited could read up to limit from fileFrom.
	fileFromLimited := io.LimitReader(fileFrom, limit)

	// Create a progress bar with limit.
	barLimit := progressBarLimit(fInfo.Size(), limit, offset)
	bar := pb.Full.Start64(barLimit)
	defer func() {
		bar.SetCurrent(barLimit)
		bar.Finish()
	}()

	// barReader is a proxy reader to calculate progress.
	barReader := bar.NewProxyReader(fileFromLimited)

	// Copy with limit via progress bar proxy or return the error.
	_, err = io.Copy(fileTo, barReader)
	if err != nil {
		return fmt.Errorf("error while copying file: %w", err)
	}

	return nil
}

// progressBarLimit returns limit for progress bar.
func progressBarLimit(filesize int64, limit int64, offset int64) int64 {
	barLimit := min(limit, filesize-offset)
	if filesize == 0 {
		barLimit = limit
	}

	return barLimit
}
