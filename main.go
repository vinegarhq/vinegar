package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

var Version string

func usage() {
	fmt.Fprintln(os.Stderr, "usage: vinegar [delete|edit|exec|kill|printconfig|reset|version]")
	fmt.Fprintln(os.Stderr, "       vinegar [player|studio] [args...]")

	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	if Config.Log {
		logOutput := io.MultiWriter(os.Stderr, LogFile("vinegar"))
		log.SetOutput(logOutput)
	}

	switch os.Args[1] {
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
	case "player":
		RobloxLaunch("RobloxPlayerLauncher.exe", "Client", os.Args[2:]...)
	case "studio":
		RobloxLaunch("RobloxStudioLauncherBeta.exe", "Studio", os.Args[2:]...)
	case "reset":
		EdgeDirSet(DirMode, false)
		DeleteDirs(Dirs.Pfx, Dirs.Log)
		CheckDirs(DirMode, Dirs.Pfx, Dirs.Log)
	case "printconfig":
		fmt.Printf("%+v\n", Config)
	case "version":
		fmt.Println("Vinegar", Version)
	default:
		usage()
	}
}
