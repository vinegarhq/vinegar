package main

import (
	"fmt"
	"os"
	"os/exec"
)

var Version = "no version set :("

func usage() {
	fmt.Fprintln(os.Stderr, "usage: vinegar [config|delete|edit|exec|kill|logs|version]")
	fmt.Fprintln(os.Stderr, "       vinegar [player|studio] [args...]")

	os.Exit(1)
}

func execWine(args ...string) {
	PfxInit()
	execCmd := exec.Command("wine", args...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	fmt.Println(execCmd.Run())
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "config":
		Config.Print()
	case "delete":
		DeleteDirs(Dirs.Data, Dirs.Cache)
	case "edit":
		EditConfig()
	case "exec":
		execWine(os.Args[2:]...)
	case "kill":
		PfxKill()
	case "logs":
		ListLogFiles()
	case "player":
		LogToFile()
		RobloxPlayer(os.Args[2:]...)
	case "studio":
		LogToFile()
		RobloxStudio(os.Args[2:]...)
	case "version":
		fmt.Println("Vinegar", Version)
	default:
		usage()
	}
}
