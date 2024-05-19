package main

import (
	"fmt"
	"os"
	"strconv"
	"os/exec"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	command := os.Args[3]
	args := os.Args[4:len(os.Args)]
	
	isError := false
	exitCode := 0
	var err error
	if args[0] == "echo_stderr" {
		args[0] = "echo"
		isError = true
	}
	if (args[0] == "exit") {
		exitCode, err = strconv.Atoi(args[1])
		if (err != nil) {
			exitCode = 0
		}
		os.Exit(exitCode)
	}
	
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Err: %v", err)
		os.Exit(1)
	}

	if (isError) {
		print(string(output)) //print prints to stderr
	} else {
		fmt.Print(string(output))
	}
	
}
