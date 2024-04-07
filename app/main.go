package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {

	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	// Create a chroot directory
	err := os.Mkdir("/tmp/my-docker-container-fs", 0755)
	if err != nil {
		fmt.Println("error creating chroot dir", err)
		os.Exit(1)
	}

	syscall.Chroot("/tmp/my-docker-container-fs")
	syscall.Chdir("/") // set the working directory inside container

	files, err := ioutil.ReadDir("/")
	if err != nil {
		fmt.Println("error reading dir", err)
	}

	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())
	}

	// Create /dev/null inside the container chroot
	// we're doing this because some commands, like the cmd.Run might expect /dev/null to exist
	err = os.Mkdir("/dev/null", 0755)
	if err != nil {
		fmt.Println("Errr", err)
		os.Exit(1)
	}

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		os.Exit(cmd.ProcessState.ExitCode())
	}
}
