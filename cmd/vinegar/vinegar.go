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

	cmd := os.Args[1]
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	switch cmd {
	// These commands don't require a configuration
	case "delete", "edit", "uninstall", "version":
		switch cmd {
		case "delete":
			Delete()
		case "edit":
			editor.EditConfig()
		case "uninstall":
			Uninstall()
		case "version":
			fmt.Println(Version)
		}
	// These commands (except player & studio) don't require a configuration,
	// but they require a wineprefix, hence wineroot of configuration is required.
	case "player", "studio", "exec", "kill", "install-webview2":
		cfg, err := config.Load()
		if err != nil {
			log.Fatal(err)
		}

		pfx := wine.New(dirs.GetPrefixPath(""))
		pfx.Interrupt()

		if err := pfx.Setup(); err != nil {
			log.Fatal(err)
		}

		switch cmd {
		case "exec":
			if err := pfx.ExecWine(os.Args[2:]...); err != nil {
				log.Fatal(err)
			}
		case "kill":
			pfx.Kill()
		case "install-webview2":
			if err := InstallWebview2(&pfx); err != nil {
				log.Fatal(err)
			}

		case "player", "studio":
			logFile := logs.File(cmd)
			logOutput := io.MultiWriter(logFile, os.Stderr)

			pfx.Output = logOutput
			log.SetOutput(logOutput)

			defer logFile.Close()

			switch cmd {
			case "player":
				Binary(roblox.Player, &cfg, &pfx, os.Args[2:]...)
			case "studio":
				Binary(roblox.Studio, &cfg, &pfx, os.Args[2:]...)
			}
		}
	default:
		usage()
	}
}

func Uninstall() {
	pfx := wine.New(dirs.GetPrefixPath(""))

	vers, err := state.Versions(&pfx)
	if err != nil {
		log.Fatal(err)
	}

	for _, ver := range vers {
		log.Println("Removing version directory", ver)

		err = os.RemoveAll(filepath.Join(dirs.GetVersionsPath(&pfx), ver))
		if err != nil {
			log.Fatal(err)
		}
	}

	err = state.ClearApplications(&pfx)
	if err != nil {
		log.Fatal(err)
	}
}

func Delete() {
	var prefix = dirs.GetPrefixPath("")

	log.Println("Deleting Wineprefix")
	if err := os.RemoveAll(prefix); err != nil {
		log.Fatal(err)
	}
}
