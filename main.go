package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
)

var Version string

func usage() {
	fmt.Fprintln(os.Stderr, "usage: vinegar [delete|edit|exec|init|kill|printconfig|reset|rfpsu|version]")
	fmt.Fprintln(os.Stderr, "       vinegar [player|studio] [args...]")
	fmt.Fprintln(os.Stderr, "       vinegar [dxvk] install|uninstall")

	os.Exit(1)
}

func SigIntInit() {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)

		<-sigChan
		log.Println("Received interrupt signal")
		PfxKill()
		os.Exit(130)
	}()
}

func LogInit() {
	if Config.Log {
		logOutput := io.MultiWriter(os.Stderr, LogFile("vinegar"))
		log.SetOutput(logOutput)
	}
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	LogInit()
	SigIntInit()

	switch os.Args[1] {
	case "delete":
		EdgeDirSet(DirMode, false)
		DeleteDirs(Dirs.Data, Dirs.Cache)
	case "dxvk":
		if len(os.Args) < 3 {
			usage()
		}

		switch os.Args[2] {
		case "install":
			DxvkInstall(true)
		case "uninstall":
			DxvkUninstall(true)
		default:
			usage()
		}
	case "edit":
		if err := EditConfig(); err != nil {
			log.Fatal("failed to edit config:", err)
		}
	case "exec":
		if err := Exec("wine", false, os.Args[2:]...); err != nil {
			log.Fatal("exec err:", err)
		}
	case "init":
		PfxInit()
	case "kill":
		PfxKill()
	case "player":
		RobloxLaunch("RobloxPlayerLauncher.exe", "Client", os.Args[2:]...)
		CommLoop("RobloxPlayerBet")
		PfxKill()
	case "studio":
		RobloxLaunch("RobloxStudioLauncherBeta.exe", "Studio", os.Args[2:]...)
		CommLoop("RobloxStudioBet")
		PfxKill()
	case "reset":
		EdgeDirSet(DirMode, false)
		DeleteDirs(Dirs.Pfx, Dirs.Log)
		CheckDirs(DirMode, Dirs.Pfx, Dirs.Log)
	case "printconfig":
		fmt.Printf("%+v\n", Config)
	case "rfpsu":
		RbxFpsUnlocker()
	case "version":
		fmt.Println("vinegar version:", Version)
	default:
		usage()
	}
}
