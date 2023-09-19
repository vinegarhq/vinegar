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
	case "edit", "version":
		switch cmd {
		case "edit":
			editor.EditConfig()
		case "version":
			fmt.Println(Version)
		}
	// These commands (except player & studio) don't require a configuration,
	// but they require a wineprefix, hence wineroot of configuration is required.

	// Note for maintainers: uninstall and delete now require a configuration...
	// ...or else they might delete the wrong wineprefix while WineHQReportMode is enabled.
	// Should the structure of this case block be changed, or is it fine the way it is?
	case "player", "studio", "exec", "kill", "install-webview2", "delete", "uninstall":
		cfg, err := config.Load()

		//Use a separate wineprefix for WineHQ reporting; this ensures that vinegar leaves no traces which might affect the results behind
		//Note to maintainers: Does this belong in config.go (like the Wine Root override) instead?
		if cfg.WineHQReportMode {
			dirs.Prefix = filepath.Join(dirs.Data, "prefix-winehq-report-mode")
			dirs.PrefixData = filepath.Join(dirs.Prefix, "vinegar")
			dirs.Versions = filepath.Join(dirs.PrefixData, "versions")

			log.Printf("WineHQReportMode: Overrding wineprefix path. Temporary prefix is located at: %s", dirs.Prefix)
		}

		if err != nil {
			log.Fatal(err)
		}

		pfx := wine.New(dirs.Prefix)
		pfx.Interrupt()

		if err := pfx.Setup(); err != nil {
			log.Fatal(err)
		}

		switch cmd {
		case "uninstall":
			Uninstall()
		case "delete":
			Delete()
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
