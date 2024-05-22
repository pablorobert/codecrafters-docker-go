package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type JwtToken struct {
	Token       string    `json:"token"`
	AccessToken string    `json:"access_token"`
	ExpiresIn   int       `json:"expires_in"`
	IssuedAt    time.Time `json:"issued_at"`
}

type Manifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	Name          string `json:"name"`
	Tag           string `json:"tag"`
	Architecture  string `json:"architecture"`
	FSLayers      []struct {
		BlobSum string `json:"blobSum"`
	} `json:"fsLayers"`
	Signatures []interface{} `json:"signatures"`
}

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	image := "library/" + os.Args[2]

	requestURL := "https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + image + ":pull"
	res, err := http.Get(requestURL)
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	var jwtToken JwtToken
	json.Unmarshal(resBody, &jwtToken)

	newDir, err := os.MkdirTemp("/tmp", "docker")
	if err != nil {
		fmt.Println("erro ao criar docker tmp dir")
		fmt.Printf("%v", err)
	}
	var Bearer = "Bearer " + jwtToken.AccessToken

	requestURL = "https://registry-1.docker.io/v2/" + image + "/manifests/latest"

	req, err := http.NewRequest("GET", requestURL, nil)
	req.Header.Add("Authorization", Bearer)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.list.v1+json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Erro lendo manifest -", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Erro lendo corpo da resposta:", err)
	}

	var manifest Manifest
	json.Unmarshal(body, &manifest)

	for layer := range manifest.FSLayers {
		blobSum := manifest.FSLayers[layer].BlobSum

		downloadBlob(image, blobSum, jwtToken.AccessToken, newDir /*, ch*/)
	}

	defer os.RemoveAll(newDir)

	err = os.MkdirAll(newDir+"/usr/bin", os.ModePerm)
	if err != nil {
		fmt.Println("erro ao criar diretorio /usr/bin")
		fmt.Printf("%v", err)
	}
	if err := os.MkdirAll(newDir+"/bin/", os.ModePerm); err != nil {
		log.Fatal(err)
	}
	if err != nil {
		fmt.Println("erro ao criar diretorio /bin ")
		fmt.Printf("%v", err)
	}

	//workaround
	err = os.MkdirAll(newDir+"/dev/null", os.ModePerm)
	if err != nil {
		fmt.Println("erro ao criar diretorio /dev/null")
		fmt.Printf("%v", err)
	}

	command = os.Args[3]
	args = os.Args[4:len(os.Args)]

	syscall.Chroot(newDir)
	os.Chdir("/")

	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		//Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS,
	}

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(2)
	}

	fmt.Print(string(output))
}

func downloadBlob(image string, blobSum string, token string, newDir string /*, ch chan int*/) {
	requestURL := "https://registry-1.docker.io/v2/" + image + "/blobs/" + blobSum

	req, err := http.NewRequest("GET", requestURL, nil)
	var Bearer = "Bearer " + token
	req.Header.Add("Authorization", Bearer)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error - ", err)
	}
	defer resp.Body.Close()

	out, err := os.Create("/tmp/layer")

	check(err)
	defer out.Close()

	out.ReadFrom(resp.Body)

	_, err = exec.Command("tar", "xf", "/tmp/layer", "-C", newDir).Output()
	if err != nil {
		fmt.Printf("tar => %+v", err)
	}
	os.RemoveAll("/tmp/layer")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
