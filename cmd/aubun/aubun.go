package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/vinegarhq/aubun/internal/config"
	"github.com/vinegarhq/aubun/internal/config/state"
	"github.com/vinegarhq/aubun/internal/dirs"
	"github.com/vinegarhq/aubun/internal/logs"
	"github.com/vinegarhq/aubun/roblox"
	"github.com/vinegarhq/aubun/roblox/bootstrapper"
	"github.com/vinegarhq/aubun/wine"
	"github.com/vinegarhq/aubun/wine/dxvk"
)

func usage() {
	fmt.Fprintln(os.Stderr, "usage: aubun kill|player|studio|exec [args...]")

	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cmd := os.Args[1]
	pfx := wine.New(dirs.Prefix, "")
	pfx.Interrupt()

	switch cmd {
	case "player", "studio":
		logFile := logs.File(cmd)
		logOutput := io.MultiWriter(logFile, os.Stderr)

		pfx.Output = logOutput
		log.SetOutput(logOutput)

		defer logFile.Close()
	}

	cfg := config.Load()
	cfg.Env.Setenv()

	if err := pfx.Setup(); err != nil {
		log.Fatal(err)
	}

	switch cmd {
	case "player":
		Binary(roblox.Player, &cfg, &pfx, os.Args[2:]...)
	case "studio":
		Binary(roblox.Studio, &cfg, &pfx, os.Args[2:]...)
	case "exec":
		if err := pfx.Exec(os.Args[2:]...); err != nil {
			log.Fatal(err)
		}
	case "kill":
		pfx.Kill()
	case "uninstall":
		Uninstall()
	case "delete":
		pfx.Kill()
		Delete()
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
	log.Println("Deleting data directory")
	if err := os.RemoveAll(dirs.Data); err != nil {
		log.Fatal(err)
	}
}

func SetupBinary(ver roblox.Version, dir string) {
	if err := dirs.Mkdirs(dir, dirs.Downloads); err != nil {
		log.Fatal(err)
	}

	manifest, err := bootstrapper.FetchManifest(ver, dirs.Downloads)
	if err != nil {
		log.Fatal(err)
	}

	if err := manifest.Download(); err != nil {
		log.Fatal(err)
	}

	if err := manifest.Extract(dir, bootstrapper.Directories(ver.Type)); err != nil {
		log.Fatal(err)
	}

	if err := state.SaveManifest(&manifest); err != nil {
		log.Fatal(err)
	}

	if err := state.CleanPackages(); err != nil {
		log.Fatal(err)
	}

	if err := state.CleanVersions(); err != nil {
		log.Fatal(err)
	}
}

func Binary(bt roblox.BinaryType, cfg *config.Config, pfx *wine.Prefix, args ...string) {
	var appCfg config.Application
	var ver roblox.Version
	name := bt.String()

	switch bt {
	case roblox.Player:
		appCfg = cfg.Player
	case roblox.Studio:
		appCfg = cfg.Studio
	}

	dxvkInstalled, err := state.DxvkInstalled()
	if err != nil {
		log.Fatal(err)
	}

	if err := dirs.Mkdirs(dirs.Cache); err != nil {
		log.Fatal(err)
	}

	if appCfg.Dxvk {
		if !dxvkInstalled {
			if err := dxvk.Fetch(dirs.Cache); err != nil {
				log.Fatal(err)
			}

			if err := dxvk.Extract(dirs.Cache, pfx); err != nil {
				log.Fatal(err)
			}

			if err := state.SaveDxvk(true); err != nil {
				log.Fatal(err)
			}
		}

		dxvk.Setenv()
	} else if dxvkInstalled && !cfg.Player.Dxvk && !cfg.Studio.Dxvk {
		if err := dxvk.Remove(pfx); err != nil {
			log.Fatal(err)
		}

		if err := state.SaveDxvk(false); err != nil {
			log.Fatal(err)
		}
	}

	channel := appCfg.Channel
	if strings.HasPrefix(strings.Join(args, " "), "roblox-player:1") {
		args, channel = bootstrapper.ParsePlayerURI(args[0])
	}

	if channel != appCfg.Channel && channel != "" {
		log.Printf("Roblox is attempting to set channel to %s from launch URI, ignoring", channel)
		channel = appCfg.Channel
	}

	if appCfg.ForcedVersion != "" {
		log.Printf("WARNING: using forced version: %s", appCfg.ForcedVersion)

		ver, err = roblox.ForceVersion(bt, channel, appCfg.ForcedVersion)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		ver, err = roblox.LatestVersion(bt, channel)
		if err != nil {
			log.Fatal(err)
		}
	}

	verDir := filepath.Join(dirs.Versions, ver.GUID)

	_, err = os.Stat(filepath.Join(verDir, "AppSettings.xml"))
	if err != nil {
		log.Printf("Updating/Installing %s", name)
		SetupBinary(ver, verDir)
	} else {
		log.Printf("%s is up to date (%s)", name, ver.GUID)
	}

	appCfg.Env.Setenv()

	err = appCfg.FFlags.SetRenderer(appCfg.Renderer)
	if err != nil {
		log.Fatal(err)
	}

	err = appCfg.FFlags.Apply(verDir)
	if err != nil {
		log.Fatal(err)
	}

	err = dirs.OverlayDir(verDir)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Launching %s", name)

	pfx.Launcher = cfg.Launcher
	args = append([]string{filepath.Join(verDir, bt.Executable())}, args...)

	if err := pfx.Exec(args...); err != nil {
		log.Fatal(err)
	}

	if appCfg.AutoKillPrefix {
		pfx.Kill()
	}
}
