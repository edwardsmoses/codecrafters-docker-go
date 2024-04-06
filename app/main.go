package main

import (
	"fmt"
	"strconv"

	// Uncomment this block to pass the first stage!
	"os"
	"os/exec"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.

	// Uncomment this block to pass the first stage!
	//

	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	if args[0] == "exit" {
		exitCode, _ := strconv.Atoi(args[1])
		os.Exit(exitCode)
	}

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		fmt.Printf("Err: %v", err)
		os.Exit(1)
	}
}
