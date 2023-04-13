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

	switch os.Args[1] {
	case "config":
		printConfig()
	case "configfile":
		printConfigFile()
	case "delete":
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
		fmt.Println(LatestLogFile("exec-*"))
		fmt.Println(LatestLogFile("vinegar-*"))
	case "player":
		LogToFile()
		RobloxPlayer(os.Args[2:]...)
	case "reset":
		DeleteDirs(Dirs.Pfx, Dirs.Log)
		CreateDirs(Dirs.Pfx, Dirs.Log)
	case "studio":
		LogToFile()
		RobloxStudio(os.Args[2:]...)
	case "version":
		fmt.Println("Vinegar", Version)
	case "print":
		fmt.Println(Config.FFlags)
	case "get":
		var pkgmanif PackageManifest

		pkgmanif.Version = GetLatestVersion()
		pkgmanif.Construct()
		pkgmanif.DownloadAll()
		pkgmanif.VerifyAll()
		pkgmanif.ExtractAll(ClientPackageDirectories())
		//			ExtractPackage("cache", "bleh", pkg)
		//		}
	default:
		usage()
	}
}
