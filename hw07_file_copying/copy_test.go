package main

import (
	"errors"
	"io/fs"
	"os"
	"testing"
)

// setupTest creates new temp file with provided size and returns path to the file.
// Ensures created file has proper size.
//
// If filesize is negative, setupTest just returns temp filename.
func setupTest(t *testing.T, filesize int) string {
	t.Helper()

	file, err := os.CreateTemp(os.TempDir(), "hw07")
	if err != nil {
		t.Fatalf("can't create temp file: %s", err)
		return ""
	}
	t.Logf("created file %s", file.Name())

	if filesize >= 0 {
		b := make([]byte, filesize)
		n, err := file.Write(b)
		if err != nil {
			t.Fatalf("can't write temp file: %s", err)
			return ""
		}
		t.Logf("... written %d bytes to %s", n, file.Name())

		if n != filesize {
			t.Fatalf("wrong file size, written %d expected %d", n, filesize)
			return ""
		}
	}

	err = file.Close()
	if err != nil {
		t.Fatalf("can't close file: %s", err)
	}

	// If filesize is negative, setupTest should return filename only: remove temp file.
	if filesize < 0 {
		teardownTest(t, file.Name())
	}

	return file.Name()
}

// teardownTest tries to remove file by its filepath.
func teardownTest(t *testing.T, filepath string) {
	t.Helper()

	t.Logf("try remove file %s", filepath)
	err := os.Remove(filepath)
	if err != nil {
		t.Fatalf("can't remove file '%s': %s", filepath, err)
	}
	t.Logf("... seems removed file %s", filepath)

	_, err = os.Stat(filepath)
	if err == nil {
		t.Fatalf("can't remove file '%s': still exists", filepath)
	}

	if !os.IsNotExist(err) {
		t.Fatalf("can't remove file '%s': %s", filepath, err)
	}
}

type testArgs struct {
	fromPath string
	toPath   string
	offset   int64
	limit    int64
}

type tableTest struct {
	name     string
	setup    func(t *testing.T) testArgs
	teardown func(t *testing.T, a testArgs)
	errIs    error
	errAs    func(err error) (error, bool)
}

func genErrAs[T error](target T) func(err error) (error, bool) {
	return func(err error) (error, bool) {
		return target, errors.As(err, &target)
	}
}

func TestCopy(t *testing.T) {
	tests := []tableTest{
		{
			name: "file doesn't exist",
			setup: func(t *testing.T) testArgs {
				t.Helper()

				from := setupTest(t, -1)
				to := setupTest(t, -1)

				return testArgs{
					fromPath: from,
					toPath:   to,
					offset:   0,
					limit:    0,
				}
			},
			errAs: genErrAs(new(fs.PathError)),
		},
		{
			name: "directory",
			setup: func(t *testing.T) testArgs {
				t.Helper()

				dirname, err := os.MkdirTemp(os.TempDir(), "hw07")
				if err != nil {
					t.Fatalf("can't create tempdir: %v", err)
				}

				to := setupTest(t, -1)

				return testArgs{
					fromPath: dirname,
					toPath:   to,
					offset:   0,
					limit:    0,
				}
			},
			teardown: func(t *testing.T, a testArgs) {
				t.Helper()

				err := os.Remove(a.fromPath)
				if err != nil {
					t.Fatalf("can't remove tempdir '%s': %v", a.fromPath, err)
				}
			},
			errIs: ErrUnsupportedFile,
		},
		{
			name: "invalid offset",
			setup: func(t *testing.T) testArgs {
				t.Helper()

				const fileSize = 10
				from := setupTest(t, fileSize)
				to := setupTest(t, -1)

				return testArgs{
					fromPath: from,
					toPath:   to,
					offset:   fileSize + 1,
					limit:    0,
				}
			},
			teardown: func(t *testing.T, a testArgs) {
				t.Helper()

				teardownTest(t, a.fromPath)
			},
			errIs: ErrOffsetExceedsFileSize,
		},
		{
			name: "can't open input file for reading",
			setup: func(t *testing.T) testArgs {
				t.Helper()

				from := setupTest(t, 0)
				to := setupTest(t, -1)

				err := os.Chmod(from, 0o000)
				if err != nil {
					t.Fatalf("can't chmod 0000 for %s: %s", from, err)
				}

				return testArgs{
					fromPath: from,
					toPath:   to,
					offset:   0,
					limit:    0,
				}
			},
			teardown: func(t *testing.T, a testArgs) {
				t.Helper()

				teardownTest(t, a.fromPath)
			},
			errAs: genErrAs(new(fs.PathError)),
		},
		{
			name: "can't open output file for writing",
			setup: func(t *testing.T) testArgs {
				t.Helper()

				from := setupTest(t, 0)
				to := setupTest(t, 0)

				err := os.Chmod(to, 0o000)
				if err != nil {
					t.Fatalf("can't chmod 0000 for %s: %s", to, err)
				}

				return testArgs{
					fromPath: from,
					toPath:   to,
					offset:   0,
					limit:    0,
				}
			},
			teardown: func(t *testing.T, a testArgs) {
				t.Helper()

				teardownTest(t, a.fromPath)
				teardownTest(t, a.toPath)
			},
			errAs: genErrAs(new(fs.PathError)),
		},
		{
			name: "valid offset",
			setup: func(t *testing.T) testArgs {
				t.Helper()

				from := setupTest(t, 10)
				to := setupTest(t, -1)

				return testArgs{
					fromPath: from,
					toPath:   to,
					offset:   10,
					limit:    0,
				}
			},
			teardown: func(t *testing.T, a testArgs) {
				t.Helper()

				teardownTest(t, a.fromPath)
				teardownTest(t, a.toPath)
			},
		},
		{
			name: "link",
			setup: func(t *testing.T) testArgs {
				t.Helper()

				to := setupTest(t, -1)

				return testArgs{
					fromPath: "./testdata/link-input.txt",
					toPath:   to,
					offset:   0,
					limit:    0,
				}
			},
			teardown: func(t *testing.T, a testArgs) {
				t.Helper()

				teardownTest(t, a.toPath)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.setup(t)
			if tt.teardown != nil {
				defer tt.teardown(t, args)
			}

			err := Copy(args.fromPath, args.toPath, args.offset, args.limit)
			if tt.errIs != nil || tt.errAs != nil {
				switch {
				case err == nil:
					t.Error("wants error, but there are no")
				case tt.errIs != nil && !errors.Is(err, tt.errIs):
					t.Errorf("wants error %#v, but got %#v", tt.errIs, err)
				case tt.errAs != nil:
					if target, ok := tt.errAs(err); !ok {
						t.Errorf("wants error %#v, but got %#v", target, err)
					}
				}
			} else if err != nil {
				t.Errorf("doesn't want error, but got %#v", err)
			}
		})
	}
}
