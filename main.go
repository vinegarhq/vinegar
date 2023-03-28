package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"

	"github.com/BurntSushi/toml"
)

var Version string

func usage() {
	fmt.Fprintln(os.Stderr, "usage: vinegar [delete|edit|exec|kill|printconfig|reset|version]")
	fmt.Fprintln(os.Stderr, "       vinegar [player|studio] [args...]")

	os.Exit(1)
}

func printDebug() {
	info, ok := debug.ReadBuildInfo()

	if !ok {
		panic("ReadBuildInfo() is not okay :(")
	}

	fmt.Printf("%s %s\n\n", "Vinegar", Version)

	// pretty print, ignores empty keys
	for _, s := range info.Settings {
		if s.Value != "" {
			fmt.Printf("%s = %s\n", s.Key, s.Value)
		}
	}

	fmt.Println()
	LatestLogFiles(2)
	fmt.Println()

	if err := toml.NewEncoder(os.Stdout).Encode(Config); err != nil {
		panic(err)
	}
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
		logToFile()
		RobloxLaunch("RobloxPlayerLauncher.exe", "Client", os.Args[2:]...)
	case "studio":
		logToFile()
		RobloxStudio(os.Args[2:]...)
	case "reset":
		EdgeDirSet(DirMode, false)
		DeleteDirs(Dirs.Pfx, Dirs.Log)
		CheckDirs(DirMode, Dirs.Pfx, Dirs.Log)
	case "debug":
		printDebug()
	default:
		usage()
	}
}
