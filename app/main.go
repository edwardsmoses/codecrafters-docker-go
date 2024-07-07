package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type TokenResponse struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
	Expires     int    `json:"expires_in"`
	IssuedAt    string `json:"issued_at"`
}
type ManiFest struct {
	Name     string     `json:"name"`
	Tag      string     `json:"tag"`
	FSLayers []fsLayers `json:"fsLayers"`
}
type fsLayers struct {
	BlobSum string `json:"blobSum"`
}


// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {

	img := os.Args[2]
	split := strings.Split(img, ":")
	repo := "library"
	image := split[0]
	tag := "latest"
	if len(split) == 2 {
		tag = split[1]
	}
	request, err := http.NewRequest("GET", fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", repo+"/"+image), nil)
	if err != nil {
		fmt.Printf("ERR!! %+v", err)
	}
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(request)
	var result TokenResponse
	json.NewDecoder(resp.Body).Decode(&result)
	// fmt.Printf("\n\nTOKEN => %+v\n\n", result.Token)
	manifestReq, err := http.NewRequest("GET", fmt.Sprintf("https://registry.hub.docker.com/v2/%s/manifests/%s", repo+"/"+image, tag), nil)
	if err != nil {
		fmt.Printf("ERR!! %+v", err)
	}
	manifestReq.Header.Add("Authorization", "Bearer "+strings.TrimSpace(result.Token))
	manifestReq.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v1+json")
	mani, err := http.DefaultClient.Do(manifestReq)
	if err != nil {
		fmt.Printf("ERRRR => %+v", err)
	}
	var manifest ManiFest
	json.NewDecoder(mani.Body).Decode(&manifest)


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

	for _, value := range manifest.FSLayers {
		req, err := http.NewRequest("GET", "https://registry-1.docker.io/v2/library/"+image+"/blobs/"+value.BlobSum, nil)
		if err != nil {
			fmt.Println("er1")
		}
		req.Header.Add("Authorization", "Bearer "+strings.TrimSpace(result.Token))
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("er2")
		}
		defer resp.Body.Close()
		f, e := os.Create(chrootDir + "/output")
		if e != nil {
			panic(e)
		}
		defer f.Close()
		f.ReadFrom(resp.Body)
		_, err = exec.Command("tar", "xf", chrootDir + "/output", "-C", chrootDir).Output()
		if err != nil {
			fmt.Printf("OUT ERR untar => %+v", err)
		}
		// fmt.Printf("output => %+v", out)
		os.RemoveAll(chrootDir + "/output")
	}

	syscall.Chroot(chrootDir)
	syscall.Chdir("/") // set the working directory inside container

	cmd := exec.Command(command, args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID,
	}

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
