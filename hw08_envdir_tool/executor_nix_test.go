//go:build !windows

package main

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// createTempFile creates new temp file with provided content and returns path to the file.
// Ensures created file has proper size.
func createTempFile(t *testing.T, content []byte) string {
	t.Helper()

	// Create temp file
	file, err := os.CreateTemp(os.TempDir(), "hw08")
	if err != nil {
		t.Fatalf("can't create temp file: %s", err)
		return ""
	}
	t.Logf("created file %s", file.Name())

	// Write content
	n, err := file.Write(content)
	if err != nil {
		t.Fatalf("can't write temp file: %s", err)
		return ""
	}
	t.Logf("... written %d bytes to %s", n, file.Name())

	// Check if all content has been written.
	if n != len(content) {
		t.Fatalf("wrong file size, written %d expected %d", n, len(content))
		return ""
	}

	err = file.Close()
	if err != nil {
		t.Fatalf("can't close file: %s", err)
	}

	return file.Name()
}

// removeFile tries to remove file by its filepath.
func removeFile(t *testing.T, filepath string) {
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

// hasExe naive checks is file could be executed.
func hasExe(filepath string) bool {
	stat, err := os.Stat(filepath)
	// May be not exists or some permission error.
	if err != nil {
		return false
	}

	return stat.Mode().Perm()&0x111 != 0
}

func TestRunCmd_Simple(t *testing.T) {
	t.Run("echo $VAR", func(t *testing.T) {
		exe := "/bin/sh"
		if !hasExe(exe) {
			t.Skipf("%s not executable", exe)
		}

		if !hasExe("/bin/echo") {
			t.Skip("/bin/echo not executable")
		}

		script := createTempFile(t, []byte(`#!/bin/sh
/bin/echo -n $VAR
      `))
		defer removeFile(t, script)

		stdout := createTempFile(t, []byte{})
		defer removeFile(t, stdout)

		stdoutFile, err := os.OpenFile(stdout, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
		if err != nil {
			t.Fatalf("os.OpenFile(stdout): %v", err)
		}

		osStdout := os.Stdout

		os.Stdout = stdoutFile
		code, err := RunCmd(
			[]string{exe, script},
			Environment{"VAR": {"var", false}},
		)
		os.Stdout = osStdout
		stdoutFile.Close()

		if err != nil {
			t.Errorf("RunCmd failed: %d: %#v", code, err)
		}

		readStdoutFile, err := os.Open(stdout)
		if err != nil {
			t.Fatalf("os.Open(stdout): %v", err)
		}
		defer readStdoutFile.Close()

		content, err := io.ReadAll(readStdoutFile)
		if err != nil {
			t.Fatalf("io.ReadAll(readStdoutFile): %v", err)
		}

		require.Equal(t, "var", string(content), "content must be equal to $VAR")
	})
}

func TestRunCmd_Complex(t *testing.T) {
	const (
		toRemove    = "HW08_REMOVE"
		toRemoveVal = "MUST BE DELETED"

		toOverwrite    = "HW08_OVERWRITE"
		toOverwriteOld = "MUST BE OVERWRITTEN"
		toOverwriteNew = "NEW VALUE"

		newVar    = "HW08_NEW"
		newVarVal = "NEW VAR"
	)

	os.Setenv(toRemove, toRemoveVal)
	os.Setenv(toOverwrite, toOverwriteOld)

	t.Run("echo $HW08-REMOVE $HW08-OVERWRITE $HW08-NEW", func(t *testing.T) {
		exe := "/bin/sh"
		if !hasExe(exe) {
			t.Skipf("%s not executable", exe)
		}

		if !hasExe("/bin/echo") {
			t.Skip("/bin/echo not executable")
		}

		script := createTempFile(t, []byte(`#!/bin/sh
/bin/echo -n "$HW08_REMOVE $HW08_OVERWRITE $HW08_NEW"
      `))
		defer removeFile(t, script)

		stdout := createTempFile(t, []byte{})
		defer removeFile(t, stdout)

		stdoutFile, err := os.OpenFile(stdout, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
		if err != nil {
			t.Fatalf("os.OpenFile(stdout): %v", err)
		}

		osStdout := os.Stdout

		os.Stdout = stdoutFile
		code, err := RunCmd(
			[]string{exe, script},
			Environment{
				toRemove:    {toRemoveVal, true},
				toOverwrite: {toOverwriteNew, false},
				newVar:      {newVarVal, false},
			},
		)
		os.Stdout = osStdout
		stdoutFile.Close()

		if err != nil {
			t.Errorf("RunCmd failed: %d: %#v", code, err)
		}

		readStdoutFile, err := os.Open(stdout)
		if err != nil {
			t.Fatalf("os.Open(stdout): %v", err)
		}
		defer readStdoutFile.Close()

		content, err := io.ReadAll(readStdoutFile)
		if err != nil {
			t.Fatalf("io.ReadAll(readStdoutFile): %v", err)
		}

		require.Equal(
			t,
			" "+toOverwriteNew+" "+newVarVal,
			string(content),
			"content must be equal to $VAR",
		)
	})
}

func Test_prepareCmd(t *testing.T) {
	const (
		toRemove    = "HW08_REMOVE"
		toRemoveVal = "MUST BE DELETED"

		toOverwrite    = "HW08_OVERWRITE"
		toOverwriteOld = "MUST BE OVERWRITTEN"
		toOverwriteNew = "NEW VALUE"

		newVar    = "HW08_NEW"
		newVarVal = "NEW VAR"
	)

	os.Setenv(toRemove, toRemoveVal)
	os.Setenv(toOverwrite, toOverwriteOld)

	t.Run("set vars", func(t *testing.T) {
		execCmd := prepareCmd(
			[]string{"test"},
			Environment{
				toRemove:    {toRemoveVal, true},
				toOverwrite: {toOverwriteNew, false},
				newVar:      {newVarVal, false},
			},
		)

		require.Contains(
			t,
			execCmd.Env,
			toOverwrite+"="+toOverwriteOld,
			"must contains "+toOverwriteOld)

		require.Contains(
			t,
			execCmd.Env,
			toOverwrite+"="+toOverwriteNew,
			"must contains "+toOverwriteNew)

		require.Contains(
			t,
			execCmd.Env,
			newVar+"="+newVarVal,
			"must contains "+newVarVal)

		require.NotContains(
			t,
			execCmd.Env,
			toRemove+"="+toRemoveVal,
			"must not contains "+toRemove)
	})

	t.Run("set no vars", func(t *testing.T) {
		execCmd := prepareCmd(
			[]string{"test"},
			Environment{},
		)

		require.NotContains(
			t,
			execCmd.Env,
			toOverwrite+"="+toOverwriteNew,
			"must not contains "+toOverwriteNew)

		require.NotContains(
			t,
			execCmd.Env,
			newVar+"="+newVarVal,
			"must not contains "+newVarVal)

		require.Contains(
			t,
			execCmd.Env,
			toRemove+"="+toRemoveVal,
			"must contains "+toRemove)
	})
}
