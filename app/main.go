package main

import (
	"fmt"
	"os"
	"syscall"
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
	if args[0] == "exit" {
		exitCode, err = strconv.Atoi(args[1])
		if (err != nil) {
			exitCode = 0
		}
		os.Exit(exitCode)
	}
	if args[0] == "ls" {
		err := syscall.Mkdir(args[1] + "/dev", 0777)
		if (err != nil) {
			fmt.Println("erro ao criar diretorio1")
			fmt.Printf("%v", err)
		}
		err = syscall.Mkdir(args[1] + "/dev/null", 0777)
		if (err != nil) {
			fmt.Println("erro ao criar diretorio2")
			fmt.Printf("%v", err)
		}
		syscall.Chroot(args[1])
		syscall.Chdir("/")
	}
	
	cmd := exec.Command(command, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID, 
    }
	output, err := cmd.Output()
	if err != nil {
		fmt.Print("No such file or directory")
		os.Exit(2)
	}

	if (isError) {
		print(string(output)) //print prints to stderr
	} else {
		fmt.Print(string(output))
	}
	
}
