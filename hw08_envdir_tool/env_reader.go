package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
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

	dirFS := os.DirFS(dir)
	for _, entry := range entries {
		// Skip dirs.
		if entry.IsDir() {
			continue
		}

		// Skip wrong file names: file should not contain '='.
		if strings.IndexByte(entry.Name(), '=') != -1 {
			continue
		}

		// Read the line from current file.
		line, err := readFileLine(dirFS, entry.Name())
		if err != nil {
			return nil, err
		}

		// Line is empty line: remove the environment variable.
		if len(line) == 0 || line[0] == '\n' {
			env[entry.Name()] = EnvValue{
				Value:      "",
				NeedRemove: true,
			}

			continue
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

// readFileLine tries to open specified filename from dirFS and read the first line.
// Returns read line or occurred error.
func readFileLine(dirFS fs.FS, filename string) ([]byte, error) {
	// Try open the file
	file, err := dirFS.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("can't open file '%s': %w", filename, err)
	}
	defer file.Close()

	// Read the first line.
	buf := bufio.NewReader(file)
	line, err := buf.ReadBytes('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("can't read file '%s': %w", filename, err)
		}
	}

	return line, nil
}
