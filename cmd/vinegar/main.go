// Copyright vinegar-development 2023

package main

import (
	"fmt"
	"os"

	"github.com/vinegar-dev/vinegar"
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

	vinegar.CheckDirs(vinegar.Dirs.Log, vinegar.Dirs.Pfx)
	vinegar.LoadConfig()

	switch args[0] {
	case "delete":
		vinegar.DeleteDirs(vinegar.Dirs.Data, vinegar.Dirs.Cache)
	case "exec":
		vinegar.Exec("wine", arg)
	case "kill":
		vinegar.PfxKill()
	case "player":
		vinegar.RobloxLaunch("RobloxPlayerLauncher.exe", PLAYERURL, true, args[1:]...)
	case "studio":
		vinegar.RobloxLaunch("RobloxStudioLauncherBeta.exe", STUDIOURL, false, args[1:]...)
	case "reset":
		vinegar.DeleteDirs(vinegar.Dirs.Pfx, vinegar.Dirs.Log)
		// Automatic creation of the directories after it has been deleted
		vinegar.CheckDirs(vinegar.Dirs.Pfx, vinegar.Dirs.Log)
	default:
		usage()
	}
}
