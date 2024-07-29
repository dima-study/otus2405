package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("can't read dir: %w", err)
	}

	env := Environment{}

	fs := os.DirFS(dir)
ENTRY:
	for _, entry := range entries {
		// Skip dirs.
		if entry.IsDir() {
			continue ENTRY
		}

		// Skip wrong file names: file should not contain '='.
		if strings.IndexByte(entry.Name(), '=') != -1 {
			continue ENTRY
		}

		// Try open the file...
		file, err := fs.Open(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("can't open file '%s': %w", entry.Name(), err)
		}

		// ... and read the first line.
		buf := bufio.NewReader(file)
		line, err := buf.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("can't read file '%s': %w", entry.Name(), err)
			}
		}

		// Line is empty line: remove the environment variable.
		if len(line) == 0 || line[0] == '\n' {
			env[entry.Name()] = EnvValue{
				Value:      "",
				NeedRemove: true,
			}

			continue ENTRY
		}

		// Okay. Line is not empty, environment variable should be set.

		// Remove terminating space, tabs and new-line.
		line = bytes.TrimRight(line, " \t\n")

		// Replace 0x00 to '\n'.
		line = bytes.ReplaceAll(line, []byte{0x00}, []byte{'\n'})

		env[entry.Name()] = EnvValue{
			Value:      string(line),
			NeedRemove: false,
		}
	}

	return env, nil
}
