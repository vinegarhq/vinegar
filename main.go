package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

var Version string

func usage() {
	fmt.Fprintln(os.Stderr, "usage: vinegar [config|configfile|delete|edit|exec|kill|logs|reset|version]")
	fmt.Fprintln(os.Stderr, "       vinegar [player|studio] [args...]")

	os.Exit(1)
}

func logToFile() {
	if !Config.Log {
		return
	}

	logOutput := io.MultiWriter(os.Stderr, LogFile("vinegar"))
	log.SetOutput(logOutput)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "config":
		printConfig()
	case "configfile":
		printConfigFile()
	case "delete":
		EdgeDirSet(DirMode, false)
		DeleteDirs(Dirs.Data, Dirs.Cache)
	case "edit":
		EditConfig()
	case "exec":
		if err := Exec("wine", false, os.Args[2:]...); err != nil {
			log.Fatal("exec err:", err)
		}
	case "kill":
		PfxKill()
	case "logs":
		fmt.Println(Dirs.Log)
		LatestLogFile("exec-*")
		LatestLogFile("vinegar-*")
	case "player":
		logToFile()
		RobloxPlayer(os.Args[2:]...)
	case "reset":
		EdgeDirSet(DirMode, false)
		DeleteDirs(Dirs.Pfx, Dirs.Log)
		CheckDirs(DirMode, Dirs.Pfx, Dirs.Log)
	case "studio":
		logToFile()
		RobloxStudio(os.Args[2:]...)
	case "version":
		fmt.Println("Vinegar", Version)
	case "print":
		fmt.Println(Config.FFlags)
	default:
		usage()
	}
}
