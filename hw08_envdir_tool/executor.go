package main

import (
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
//
// Also sets Stdin, Stdout and Stderr.
func RunCmd(cmdline []string, env Environment) (returnCode int) {
	return
}

// prepareCmd creates new command and sets env environment variables from env.
func prepareCmd(cmdline []string, env Environment) *exec.Cmd {
	cmd := exec.Command(cmdline[0], cmdline[1:]...)

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
