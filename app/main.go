package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {

	chrootDir := "/tmp/my-docker-daemon-fs"

	// Ensure any previous chroot environment is cleaned up before attempting to recreate it.
	cleanup(chrootDir)

	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	// Create a chroot directory
	if err := os.Mkdir(chrootDir, 0755); err != nil {
		fmt.Println("error creating chroot dir", err)
		os.Exit(1)
	}

	//create the /dev directory
	if err := os.Mkdir(chrootDir+"/dev", 0755); err != nil {
		fmt.Println("error creating /dev dir", err)
		os.Exit(1)
	}

	// Mount /dev to the chroot directory
	if err := syscall.Mount("/dev", chrootDir+"/dev", "", syscall.MS_BIND, ""); err != nil {
		fmt.Println("error mounting /dev:", err)
		os.Exit(1)
	}

	// create the /usr directory
	if err := os.Mkdir(chrootDir+"/usr", 0755); err != nil {
		fmt.Println("error creating /usr dir", err)
		os.Exit(1)
	}

	// Mount /usr to the chroot directory
	if err := syscall.Mount("/usr", chrootDir+"/usr", "", syscall.MS_BIND, ""); err != nil {
		fmt.Println("error mounting /usr:", err)
		os.Exit(1)
	}

	syscall.Chroot(chrootDir)
	syscall.Chdir("/") // set the working directory inside container

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("error running command:", err) // Print the error to the console
		os.Exit(cmd.ProcessState.ExitCode())
	}

	// Unmount and remove the chroot directory when the program exits
	defer cleanup(chrootDir)

}

// cleanup attempts to unmount directories and remove the chroot directory.
func cleanup(dir string) {
	syscall.Unmount(dir+"/dev", 0) // Attempt to unmount, but ignore errors as it may not be mounted
	syscall.Unmount(dir+"/usr", 0) // Same as above
	if err := os.RemoveAll(dir); err != nil {
		fmt.Printf("Warning: failed to clean up chroot dir: %v\n", err)
	}
}
