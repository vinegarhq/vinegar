package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/config/editor"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/roblox"
	"golang.org/x/term"
)

var (
	BinPrefix  string
	ConfigPath string
	FirstRun   bool
	Version    string
)

func init() {
	flag.StringVar(&ConfigPath, "config", filepath.Join(dirs.Config, "config.toml"), "config.toml file which should be used")
	flag.BoolVar(&FirstRun, "firstrun", false, "to trigger first run behavior")
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: vinegar [-config filepath] [-firstrun] player|studio run [args...]")
	fmt.Fprintln(os.Stderr, "       vinegar [-config filepath] player|studio kill|winetricks")
	fmt.Fprintln(os.Stderr, "       vinegar [-config filepath] sysinfo")
	fmt.Fprintln(os.Stderr, "       vinegar delete|edit|version")
	os.Exit(1)
}

func main() {
	flag.Parse()

	cmd := flag.Arg(0)
	args := flag.Args()

	switch cmd {
	case "delete", "edit", "version":
		switch cmd {
		case "delete":
			slog.Info("Deleting Wineprefixes and Roblox Binary deployments!")

			if err := os.RemoveAll(dirs.Prefixes); err != nil {
				log.Fatalf("remove %s: %s", dirs.Prefixes, err)
			}
		case "edit":
			if err := editor.Edit(ConfigPath); err != nil {
				log.Fatalf("edit %s: %s", ConfigPath, err)
			}
		case "version":
			fmt.Println("Vinegar", Version)
		}
	case "player", "studio", "sysinfo":
		// Remove after a few releases
		if _, err := os.Stat(dirs.Prefix); err == nil {
			slog.Info("Deleting deprecated old Wineprefix!")
			if err := os.RemoveAll(dirs.Prefix); err != nil {
				log.Fatalf("delete old prefix %s: %s", dirs.Prefix, err)
			}
		}

		cfg, err := config.Load(ConfigPath)
		if err != nil {
			log.Fatalf("load config %s: %s", ConfigPath, err)
		}

		var bt roblox.BinaryType
		switch cmd {
		case "player":
			bt = roblox.Player
		case "studio":
			bt = roblox.Studio
		case "sysinfo":
			PrintSysinfo(&cfg)
			os.Exit(0)
		}

		b, err := NewBinary(bt, &cfg)
		if err != nil {
			log.Fatal(err)
		}

		switch flag.Arg(1) {
		case "exec":
			if len(args) < 2 {
				usage()
			}

			if err := b.Prefix.Wine(args[2], args[3:]...).Run(); err != nil {
				log.Fatalf("exec prefix %s: %s", bt, err)
			}
		case "kill":
			b.Prefix.Kill()
		case "winetricks":
			if err := b.Prefix.Winetricks(); err != nil {
				log.Fatalf("exec winetricks %s: %s", bt, err)
			}
		case "run":
			err = b.Main(args[2:]...)
			if err == nil {
				slog.Info("Goodbye")
				os.Exit(0)
			}

			// Only fatal print the error if we are in a terminal, otherwise
			// display a dialog message.
			if !cfg.Splash.Enabled || term.IsTerminal(int(os.Stderr.Fd())) {
				log.Fatal(err)
			}

			slog.Error(err.Error())
			b.Splash.SetMessage("Oops!")
			b.Splash.Dialog(fmt.Sprintf(DialogFailure, err), false)
			os.Exit(1)
		default:
			usage()
		}
	default:
		usage()
	}
}

func LogFile(name string) (*os.File, error) {
	if err := dirs.Mkdirs(dirs.Logs); err != nil {
		return nil, err
	}

	// name-2006-01-02T15:04:05Z07:00.log
	path := filepath.Join(dirs.Logs, name+"-"+time.Now().Format(time.RFC3339)+".log")

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s log file: %w", name, err)
	}

	slog.Info("Logging to file", "path", path)

	return file, nil
}
