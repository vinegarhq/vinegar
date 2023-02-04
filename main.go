// Copyright vinegar-development 2023

package main

import (
	"os"
	"fmt"

	"github.com/vinegar-dev/vinegar/util"
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

	if argsCount > 1 {
		arg = args[1]
	}


	dirs := util.InitDirs()
	util.DirsCheck(dirs.Log, dirs.Pfx, dirs.Exe)
	util.InitEnv(dirs)

	switch args[0] {
	case "delete":
		util.DeleteDir(dirs.Data, dirs.Cache)
	case "exec":
		util.Exec(dirs, "wine", arg)
	case "kill":
		util.PfxKill(dirs)
	case "player":
		util.RobloxLaunch(dirs, "RobloxPlayerLauncher.exe", PLAYERURL, "Roblox Player", arg)
//	case "studio":
//		studioPath := util.InitExec(dirs, "RobloxStudioLauncherBeta.exe", STUDIOURL, "Roblox Studio")
//		util.Exec(dirs, "wine", studioPath, fargs)
	case "reset":
		util.DeleteDir(dirs.Pfx, dirs.Log)
		// Automatic creation of the directories after it has been deleted
		util.DirsCheck(dirs.Pfx, dirs.Log)
	default:
		usage()
	}
}
