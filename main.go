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
	var fargs string

	args := os.Args[1:]
	argsCount := len(args)

	if argsCount < 1 {
		usage()
	}

	if argsCount > 1 {
		fargs = args[1]
	}

	dirs := util.InitDirs()
	util.DirsCheck(dirs.Log, dirs.Pfx, dirs.Exe)
	util.InitEnv(dirs)

	switch args[0] {
	case "delete":
		util.DeleteDir(dirs.Pfx)
	case "exec":
		util.Exec(dirs, "wine", fargs)
	case "kill":
		util.PfxKill(dirs)
	case "player":
		playerPath := util.InitExec(dirs, "RobloxPlayerLauncher.exe", PLAYERURL, "Roblox Player")
		// This is undocumented roblox behavior. Don't mess with its order.
		util.Exec(dirs, "wine", playerPath, fargs, "-fast")
		util.RbxFpsUnlocker(dirs)
	case "studio":
		studioPath := util.InitExec(dirs, "RobloxStudioLauncherBeta.exe", STUDIOURL, "Roblox Studio")
		util.Exec(dirs, "wine", studioPath, fargs)
	case "reset":
		util.DeleteDir(dirs.Pfx)
		// Automatic creation of the prefix after it has been deleted
		util.DirsCheck(dirs.Pfx)
	default:
		usage()
	}
}
