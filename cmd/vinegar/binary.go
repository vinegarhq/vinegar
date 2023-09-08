package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/config/state"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/roblox/bootstrapper"
	"github.com/vinegarhq/vinegar/wine"
	"github.com/vinegarhq/vinegar/wine/dxvk"
)

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

	dxvkVersion, err := state.DxvkVersion()
	if err != nil {
		log.Fatal(err)
	}
	dxvkInstalled := dxvkVersion != ""

	if err := dirs.Mkdirs(dirs.Cache); err != nil {
		log.Fatal(err)
	}

	if appCfg.Dxvk {
		dxvkPath := filepath.Join(dirs.Cache, "dxvk-"+cfg.DxvkVersion+".tar.gz")

		if !dxvkInstalled || cfg.DxvkVersion != dxvkVersion {
			if err := dxvk.Fetch(dxvkPath, cfg.DxvkVersion); err != nil {
				log.Fatal(err)
			}

			if err := dxvk.Extract(dxvkPath, pfx); err != nil {
				log.Fatal(err)
			}

			if err := state.SaveDxvk(cfg.DxvkVersion); err != nil {
				log.Fatal(err)
			}
		}

		dxvk.Setenv()
	} else if dxvkInstalled && !cfg.Player.Dxvk && !cfg.Studio.Dxvk {
		if err := dxvk.Remove(pfx); err != nil {
			log.Fatal(err)
		}

		if err := state.SaveDxvk(""); err != nil {
			log.Fatal(err)
		}
	}

	if strings.HasPrefix(strings.Join(args, " "), "roblox-studio:1") {
		args = []string{"-protocolString", args[0]}
	}

	if appCfg.ForcedVersion != "" {
		log.Printf("WARNING: using forced version: %s", appCfg.ForcedVersion)

		ver, err = roblox.NewVersion(bt, appCfg.Channel, appCfg.ForcedVersion)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		ver, err = roblox.LatestVersion(bt, appCfg.Channel)
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

	if cfg.MultipleInstances {
		mutexer := pfx.Command("wine", filepath.Join(BinPrefix, "robloxmutexer.exe"))
		err = mutexer.Start()
		if err != nil {
			log.Printf("Failed to launch robloxmutexer: %s", err)
		}
	}

	log.Printf("Launching %s", name)

	pfx.Launcher = strings.Fields(cfg.Launcher)
	args = append([]string{filepath.Join(verDir, bt.Executable())}, args...)

	if err := pfx.ExecWine(args...); err != nil {
		log.Fatal(err)
	}

	if appCfg.AutoKillPrefix {
		pfx.Kill()
	}
}
