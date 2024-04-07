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

	os.Remove("/tmp/my-docker-daemon-fs") // Remove the chroot directory

	// Create a chroot directory
	if err := os.Mkdir("/tmp/my-docker-daemon-fs", 0755); err != nil {
		fmt.Println("error creating chroot dir", err)
		os.Exit(1)
	}

	//create the /dev directory
	if err := os.Mkdir("/tmp/my-docker-daemon-fs/dev", 0755); err != nil {
		fmt.Println("error creating /dev dir", err)
		os.Exit(1)
	}

	// Mount /dev to the chroot directory
	if err := syscall.Mount("/dev", "/tmp/my-docker-daemon-fs/dev", "", syscall.MS_BIND, ""); err != nil {
		fmt.Println("error mounting /dev:", err)
		os.Exit(1)
	}

	// create the /usr directory
	if err := os.Mkdir("/tmp/my-docker-daemon-fs/usr", 0755); err != nil {
		fmt.Println("error creating /usr dir", err)
		os.Exit(1)
	}

	// Mount /usr to the chroot directory
	if err := syscall.Mount("/usr", "/tmp/my-docker-daemon-fs/usr", "", syscall.MS_BIND, ""); err != nil {
		fmt.Println("error mounting /usr:", err)
		os.Exit(1)
	}

	// Unmount the directories when the program exits
	defer syscall.Unmount("/tmp/my-docker-daemon-fs/dev", 0)
	defer syscall.Unmount("/tmp/my-docker-daemon-fs/usr", 0)
	defer os.Remove("/tmp/my-docker-daemon-fs") // Remove the chroot directory

	syscall.Chroot("/tmp/my-docker-daemon-fs")
	syscall.Chdir("/") // set the working directory inside container

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("error running command:", err) // Print the error to the console
		os.Exit(cmd.ProcessState.ExitCode())
	}

}
