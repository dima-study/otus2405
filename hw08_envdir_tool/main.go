package main

import (
	"fmt"
	"os"
)

func main() {
	// Check if args has at least 3 arguments:
	// 1. current program
	// 2. env path
	// 3. program to run
	if len(os.Args) < 3 {
		fmt.Printf("Usage:\n\t%s env_path program [program arguments]\n",
			os.Args[0],
		)
		os.Exit(1)
	}

	// Read variables or return error.
	env, err := ReadDir(os.Args[1])
	if err != nil {
		fmt.Printf("env_path error: %s\n", err)
		os.Exit(2)
	}

	// Run program with arguments and env variables.
	returnCode, err := RunCmd(os.Args[2:], env)
	if err != nil {
		fmt.Printf("program error: %s\n", err)
	}

	// Finally exit with program returned code.
	os.Exit(returnCode)
}
