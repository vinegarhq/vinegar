// Copyright vinegar-development 2023

package main

import (
	"os"
	"fmt"

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

	dirs := vinegar.InitDirs()
	vinegar.DirsCheck(dirs.Log, dirs.Pfx, dirs.Exe)
	vinegar.InitEnv(dirs)

	switch args[0] {
	case "delete":
		vinegar.DeleteDir(dirs.Data, dirs.Cache)
	case "exec":
		vinegar.Exec(dirs, "wine", arg)
	case "kill":
		vinegar.PfxKill(dirs)
	case "player":
		vinegar.RobloxLaunch(dirs, "RobloxPlayerLauncher.exe", PLAYERURL, args[1:]...)
	case "studio":
		vinegar.RobloxLaunch(dirs, "RobloxStudioLauncherBeta.exe", STUDIOURL, args[1:]...)
	case "reset":
		vinegar.DeleteDir(dirs.Pfx, dirs.Log)
		// Automatic creation of the directories after it has been deleted
		vinegar.DirsCheck(dirs.Pfx, dirs.Log)
	default:
		usage()
	}
}

