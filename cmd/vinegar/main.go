package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/state"
)

var (
	ConfigPath string
	Version    string
)

func init() {
	flag.StringVar(&ConfigPath, "config", filepath.Join(dirs.Config, "config.toml"), "config.toml file which should be used")
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: vinegar [-config filepath] [-firstrun] run|exec [args...]")
	fmt.Fprintln(os.Stderr, "       vinegar [-config filepath] sysinfo|kill|winetricks")
	fmt.Fprintln(os.Stderr, "       vinegar delete|edit|uninstall|version")
	os.Exit(1)
}

func main() {
	flag.Parse()

	ui := New()
	defer ui.Unref()

	if code := ui.app.Run(len(os.Args), os.Args); code > 0 {
		os.Exit(code)
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

	s.Studio.Version = ""
	s.Studio.Packages = nil

	if err := s.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}
