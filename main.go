// Copyright vinegar-development 2023

package main

import (
	"fmt"
	"os"
)

const (
	PLAYERURL = "https://www.roblox.com/download/client"
	STUDIOURL = "https://www.roblox.com/download/studio"
)

func usage() {
	fmt.Println("usage: vinegar [delete|exec|kill|player|reset|studio] [arg]")
	os.Exit(1)
}

func main() {
	var arg string

	args := os.Args[1:]
	argsCount := len(args)

	if argsCount < 1 {
		usage()
	}

	CheckDirs(Dirs.Log, Dirs.Pfx)

	switch args[0] {
	case "delete":
		DeleteDirs(Dirs.Data, Dirs.Cache)
	case "exec":
		Exec("wine", arg)
	case "kill":
		PfxKill()
	case "player":
		RobloxLaunch("RobloxPlayerLauncher.exe", PLAYERURL, true, args[1:]...)
	case "studio":
		RobloxLaunch("RobloxStudioLauncherBeta.exe", STUDIOURL, false, args[1:]...)
	case "reset":
		DeleteDirs(Dirs.Pfx, Dirs.Log)
		// Automatic creation of the directories after it has been deleted
		CheckDirs(Dirs.Pfx, Dirs.Log)
	default:
		usage()
	}
}
