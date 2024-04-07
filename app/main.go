package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {

	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	// Create a chroot directory
	if err := os.Mkdir("/tmp/my-docker-container-fs", 0755); err != nil {
		fmt.Println("error creating chroot dir", err)
		os.Exit(1)
	}

	if err := os.Mkdir("/tmp/my-docker-container-fs/dev", 0755); err != nil {
		fmt.Println("error creating /dev dir", err)
		os.Exit(1)
	}

	if err := syscall.Mount("/dev", "/tmp/my-docker-container-fs/dev", "", syscall.MS_BIND, ""); err != nil {
		fmt.Println("error mounting /dev:", err)
		os.Exit(1)
	}

	defer syscall.Unmount("/tmp/my-docker-container-fs/dev", 0) // Use defer to ensure unmount is called even if the program exits early.

	syscall.Chroot("/tmp/my-docker-container-fs")
	syscall.Chdir("/") // set the working directory inside container

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("error running command:", err) // Print the error to the console
		os.Exit(cmd.ProcessState.ExitCode())
	}

	// Remember to unmount in defer statement or after your operations are complete
	syscall.Unmount("/tmp/my-docker-container-fs/dev", 0)

}
