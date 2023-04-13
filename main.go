package main

import (
	"fmt"
	"log"
	"os"
)

var Version = "no version set :("

func usage() {
	fmt.Fprintln(os.Stderr, "usage: vinegar [config|delete|edit|exec|kill|logs]")
	fmt.Fprintln(os.Stderr, "       vinegar [player|studio] [args...]")

	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	log.Println("Vinegar", Version)

	switch os.Args[1] {
	case "config":
		printConfig()
	case "delete":
		DeleteDirs(Dirs.Data, Dirs.Cache)
	case "edit":
		EditConfig()
	case "exec":
		if err := Exec("wine", "", os.Args[2:]...); err != nil {
			log.Fatal(err)
		}
	case "kill":
		PfxKill()
	case "logs":
		fmt.Println(Dirs.Log)
		fmt.Println(LatestLogFile("*.log"))
		fmt.Println(LatestLogFile("vinegar*.log"))
	case "player":
		LogToFile()
		RobloxPlayer(os.Args[2:]...)
	case "studio":
		LogToFile()
		RobloxStudio(os.Args[2:]...)
	default:
		usage()
	}
}
