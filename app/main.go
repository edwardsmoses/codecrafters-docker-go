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
	err := os.Mkdir("/tmp/my-docker-container-fs", 0755)
	if err != nil {
		fmt.Println("error creating chroot dir", err)
		os.Exit(1)
	}

	// copy the root filesystem to the chroot directory
	err = exec.Command("cp", "-a", "/rootfs/.", "/tmp/my-docker-container-fs").Run()
	if err != nil {
		fmt.Println("error copying root filesystem", err)
	}
	syscall.Mount("/dev/sda1", "/tmp/my-docker-container-fs", "ext4", syscall.MS_BIND, "")

	syscall.Chroot("/tmp/my-docker-container-fs")
	syscall.Chdir("/") // set the working directory inside container

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		fmt.Println("error running command:", err) // Print the error to the console
		os.Exit(cmd.ProcessState.ExitCode())
	}

	// Unmount the root filesystem
	syscall.Unmount("/tmp/my-docker-container-fs", syscall.MNT_DETACH)
}
