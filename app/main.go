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

	// Create the /dev directory inside the chroot
	err = os.Mkdir("/tmp/my-docker-container-fs/dev", 0755)
	if err != nil {
		fmt.Println("error creating /dev dir", err)
		os.Exit(1)
	}

	// create the /dev/null file inside the chroot
	_, err = os.Create("/tmp/my-docker-container-fs/dev/null")
	if err != nil {
		fmt.Println("error creating /dev/null file", err)
		os.Exit(1)
	}

	syscall.Chroot("/tmp/my-docker-container-fs")
	syscall.Chdir("/") // set the working directory inside container

	files, err := ioutil.ReadDir("/")
	if err != nil {
		fmt.Println("error reading dir", err)
	}

	fmt.Println("Files in chroot", files)

	for _, file := range files {
		fmt.Println(file.Name(), file.IsDir())
	}

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		os.Exit(cmd.ProcessState.ExitCode())
	}
}
