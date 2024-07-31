package main

import (
	"os"
	"os/exec"
	"strings"
)

const (
	CodeStartFail = -128
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
// Also sets Stdin, Stdout and Stderr.
//
// Returns CodeStartFail and the error once command start is failed.
// Returns exit code and possible occurred error.
func RunCmd(cmdline []string, env Environment) (returnCode int, err error) {
	execCmd := prepareCmd(cmdline, env)

	// Set stdin, stdout, stderr
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// Try to start
	err = execCmd.Start()
	if err != nil {
		return CodeStartFail, err
	}

	// Wait until command completes and return exit code with possible error.
	err = execCmd.Wait()
	return execCmd.ProcessState.ExitCode(), err
}

// prepareCmd creates new command and sets env environment variables from env.
func prepareCmd(cmdline []string, env Environment) *exec.Cmd {
	cmd := exec.Command(cmdline[0], cmdline[1:]...) //nolint:gosec

	// Current env
	osEnv := os.Environ()

	// For each env var check if it should be removed and remove it.
	for i := len(osEnv) - 1; i >= 0; i-- {
		// Get env var name.
		n := strings.IndexByte(osEnv[i], '=')
		envar := osEnv[i][:n]

		// Remove when needed.
		if v, ok := env[envar]; ok && v.NeedRemove {
			copy(osEnv[i:], osEnv[i+1:])
			osEnv = osEnv[:len(osEnv)-1]
		}
	}

	// Set rest env vars.
	for k, v := range env {
		if v.NeedRemove {
			continue
		}

		osEnv = append(osEnv, k+"="+v.Value)
	}

	cmd.Env = osEnv

	return cmd
}
