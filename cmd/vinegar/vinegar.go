package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/config/editor"
	"github.com/vinegarhq/vinegar/internal/config/state"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/logs"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/wine"
)

var (
	Version   string
	BinPrefix string
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: vinegar player|studio|exec [args...]")
	fmt.Fprintln(os.Stderr, "       vinegar edit|kill|uninstall|delete|version|install-webview2")

	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cmd := os.Args[1]
	pfx := wine.New(dirs.Prefix)
	pfx.Interrupt()

	switch cmd {
	case "player", "studio":
		logFile := logs.File(cmd)
		logOutput := io.MultiWriter(logFile, os.Stderr)

		pfx.Output = logOutput
		log.SetOutput(logOutput)

		defer logFile.Close()
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	if err := pfx.Setup(); err != nil {
		log.Fatal(err)
	}

	switch cmd {
	case "player":
		Binary(roblox.Player, &cfg, &pfx, os.Args[2:]...)
	case "studio":
		Binary(roblox.Studio, &cfg, &pfx, os.Args[2:]...)
	case "edit":
		editor.EditConfig()
	case "exec":
		if err := pfx.ExecWine(os.Args[2:]...); err != nil {
			log.Fatal(err)
		}
	case "kill":
		pfx.Kill()
	case "uninstall":
		Uninstall()
	case "delete":
		pfx.Kill()
		Delete()
	case "version":
		fmt.Println(Version)
	case "install-webview2":
		if err := InstallWebview2(&pfx); err != nil {
			log.Fatal(err)
		}
	default:
		usage()
	}
}

func Uninstall() {
	vers, err := state.Versions()
	if err != nil {
		log.Fatal(err)
	}

	for _, ver := range vers {
		log.Println("Removing version directory", ver)

		err = os.RemoveAll(filepath.Join(dirs.Versions, ver))
		if err != nil {
			log.Fatal(err)
		}
	}

	err = state.ClearApplications()
	if err != nil {
		log.Fatal(err)
	}
}

func Delete() {
	log.Println("Deleting Wineprefix")
	if err := os.RemoveAll(dirs.Prefix); err != nil {
		log.Fatal(err)
	}
}
