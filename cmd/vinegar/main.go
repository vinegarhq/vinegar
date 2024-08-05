package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/apprehensions/rbxweb/clientsettings"
	"github.com/lmittmann/tint"
	"github.com/vinegarhq/vinegar/config"
	"github.com/vinegarhq/vinegar/config/editor"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/state"
)

var (
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
	fmt.Fprintln(os.Stderr, "       vinegar delete|edit|uninstall|version")
	os.Exit(1)
}

func main() {
	flag.Parse()

	cmd := flag.Arg(0)
	args := flag.Args()

	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, nil)))

	switch cmd {
	case "delete", "edit", "uninstall", "version":
		switch cmd {
		case "delete":
			if err := Delete(); err != nil {
				log.Fatal(err)
			}
		case "edit":
			if err := editor.Edit(ConfigPath); err != nil {
				log.Fatalf("edit %s: %s", ConfigPath, err)
			}
		case "uninstall":
			if err := Uninstall(); err != nil {
				log.Fatal(err)
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

		var bt clientsettings.BinaryType
		switch cmd {
		case "player":
			bt = clientsettings.WindowsPlayer
		case "studio":
			bt = clientsettings.WindowsStudio64
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
			if code := b.Main(args[2:]...); code > 0 {
				os.Exit(code)
			}
		default:
			usage()
		}
	default:
		usage()
	}
}

func Delete() error {
	slog.Info("Deleting Wineprefixes!")

	if err := os.RemoveAll(dirs.Prefixes); err != nil {
		return fmt.Errorf("remove prefixes: %w", err)
	}

	s, err := state.Load()
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	s.Player.DxvkVersion = ""
	s.Studio.DxvkVersion = ""

	if err := s.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

func Uninstall() error {
	slog.Info("Deleting Roblox Binary deployments!")

	if err := os.RemoveAll(dirs.Versions); err != nil {
		return fmt.Errorf("remove versions: %w", err)
	}

	s, err := state.Load()
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	s.Player.Version = ""
	s.Player.Packages = nil
	s.Studio.Version = ""
	s.Studio.Packages = nil

	if err := s.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}
