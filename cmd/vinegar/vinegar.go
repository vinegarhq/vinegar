package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/config/editor"
	"github.com/vinegarhq/vinegar/internal/config/state"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/logs"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/wine"
	"github.com/vinegarhq/vinegar/extra"
)

var BinPrefix string

func usage() {
	fmt.Fprintln(os.Stderr, "usage: vinegar [-config filepath] player|studio [args...]")
	fmt.Fprintln(os.Stderr, "usage: vinegar [-config filepath] exec prog [args...]")
	fmt.Fprintln(os.Stderr, "       vinegar [-config filepath] edit|kill|uninstall|delete|install-webview2|winetricks|reportinfo")
	os.Exit(1)
}

func main() {
	configPath := flag.String("config", filepath.Join(dirs.Config, "config.toml"), "config.toml file which should be used")
	flag.Parse()

	cmd := flag.Arg(0)
	args := flag.Args()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	switch cmd {
	// These commands don't require a configuration
	case "delete", "edit", "uninstall":
		switch cmd {
		case "delete":
			Delete()
		case "edit":
			editor.EditConfig(*configPath)
		case "uninstall":
			Uninstall()
		}
	// These commands (except player & studio) don't require a configuration,
	// but they require a wineprefix, hence wineroot of configuration is required.
	case "player", "studio", "exec", "kill", "install-webview2", "winetricks":
		pfxKilled := false
		cfg, err := config.Load(*configPath)
		if err != nil {
			log.Fatal(err)
		}

		pfx := wine.New(dirs.Prefix)
		// Always ensure its created, wine will complain if the root
		// directory doesnt exist
		if err := os.MkdirAll(dirs.Prefix, 0o755); err != nil {
			log.Fatal(err)
		}

		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

		go func() {
			<-c
			pfxKilled = true
			pfx.Kill()

			if pfxKilled {
				os.Exit(0)
			}
		}()

		switch cmd {
		case "exec":
			if len(args) < 2 {
				usage()
			}

			if err := pfx.Wine(args[1], args[2:]...).Run(); err != nil {
				log.Fatal(err)
			}
		case "kill":
			pfx.Kill()
		case "install-webview2":
			if err := InstallWebview2(&pfx); err != nil {
				log.Fatal(err)
			}
		case "reportinfo":
			report, err := extra.GenerateInfo(cfg)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%+v\n", report)
		case "winetricks":
			if err := pfx.Winetricks(); err != nil {
				log.Fatal(err)
			}
		case "player", "studio":
			var b Binary

			logFile := logs.File(cmd)
			logOutput := io.MultiWriter(logFile, os.Stderr)

			pfx.Output = logOutput
			log.SetOutput(logOutput)

			defer logFile.Close()

			switch cmd {
			case "player":
				b = NewBinary(roblox.Player, &cfg, &pfx)
			case "studio":
				b = NewBinary(roblox.Studio, &cfg, &pfx)
			}

			go func() {
				err := b.Splash.Run()
				if err != nil {
					log.Fatal(err)
				}
			}()

			b.Splash.Desc(b.Config.Channel)

			errHandler := func(err error) {
				if !cfg.Splash.Enabled {
					log.Fatal(err)
				}

				log.Println(err)
				b.Splash.ShowLog(logFile.Name())
				select {} // wait for window to close
			}

			if _, err := os.Stat(filepath.Join(pfx.Dir, "drive_c", "windows")); err != nil {
				log.Printf("Initializing wineprefix at %s", pfx.Dir)

				b.Splash.Message("Initializing wineprefix")
				if err := PrefixInit(&pfx); err != nil {
					b.Splash.Message(err.Error())
					errHandler(err)
				}
			}

			if err := b.Setup(); err != nil {
				b.Splash.Message("Failed to setup Roblox")
				errHandler(err)
			}

			if err := b.Run(args[1:]...); err != nil {
				b.Splash.Message("Failed to run Roblox")
				errHandler(err)
			}
		}
	default:
		usage()
	}
}

func PrefixInit(pfx *wine.Prefix) error {
	if err := pfx.Command("wineboot", "-i").Run(); err != nil {
		return err
	}

	if err := pfx.DisableCrashDialogs(); err != nil {
		return err
	}

	return pfx.RegistryAdd("HKEY_CURRENT_USER\\Control Panel\\Desktop", "LogPixels", wine.REG_DWORD, "97")
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
