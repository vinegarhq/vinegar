package main

import (
	"fmt"
	"log"
	"os"
)

var Version string

func usage() {
	fmt.Fprintln(os.Stderr, "usage: vinegar [config|configfile|delete|edit|exec|kill|logs|reset|version]")
	fmt.Fprintln(os.Stderr, "       vinegar [player|studio] [args...]")

	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	log.Println("Vinegar", Version)

	switch os.Args[1] {
	case "player":
		RobloxPlayer(os.Args[2:]...)
	case "config":
		printConfig()
	case "delete":
		DeleteDirs(Dirs.Data, Dirs.Cache)
		//	case "edit":
		//		EditConfig()
	case "exec":
		if err := Exec("wine", false, os.Args[2:]...); err != nil {
			log.Fatal(err)
		}
	case "kill":
		PfxKill()
		//	case "logs":
		//		fmt.Println(Dirs.Log)
		//		fmt.Println(LatestLogFile("exec-*"))
		//		fmt.Println(LatestLogFile("vinegar-*"))
		//	case "player":
		//		LogToFile()
		//		RobloxPlayer(os.Args[2:]...)
	case "studio":
		RobloxStudio(os.Args[2:]...)
	default:
		usage()
	}
}
