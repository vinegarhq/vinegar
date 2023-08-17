package main

import (
	"log"
	"os"
	"strings"
	"path/filepath"

	"github.com/vinegarhq/aubun/bootstrapper"
	"github.com/vinegarhq/aubun/wine"
	"github.com/vinegarhq/aubun/wine/dxvk"
	"github.com/vinegarhq/aubun/internal/config"
	"github.com/vinegarhq/aubun/internal/config/state"
	"github.com/vinegarhq/aubun/internal/dirs"
)

func Setup(ver bootstrapper.Version, dir string) {
	if err := dirs.Mkdir(dir); err != nil {
		log.Fatal(err)
	}

	if err := dirs.Mkdir(dirs.Downloads); err != nil {
		log.Fatal(err)
	}

	manifest, err := bootstrapper.FetchManifest(ver, dirs.Downloads)
	if err != nil {
		log.Fatal(err)
	}

	if err := manifest.Download(); err != nil {
		log.Fatal(err)
	}

	if err := manifest.Extract(dir, ver.Type.Directories()); err != nil {
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

func Binary(pfx *wine.Prefix, bin bootstrapper.BinaryType, cfg config.Application, allDxvk bool, args ...string) {
	ver, err := bootstrapper.LatestVersion(bin, cfg.Channel)
	if err != nil {
		log.Fatal(err)
	}
	verDir := filepath.Join(dirs.Versions, ver.GUID)

	storedVersion, err := state.Version(bin)
	if err != nil {
		log.Fatal(err)
	}

	if storedVersion != ver.GUID {
		log.Printf("Updating %s", bin.String())
		Setup(ver, verDir)
	} else {
		log.Printf("%s is up to date (%s)", bin.String(), storedVersion)
	}

	dxvkInstalled, err := state.DxvkInstalled()
	if err != nil {
		log.Fatal(err)
	}

	if cfg.Dxvk {
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
	} else if allDxvk {
		if err := dxvk.Remove(pfx); err != nil {
			log.Fatal(err)
		}

		if err := state.SaveDxvk(false); err != nil {
			log.Fatal(err)
		}
	}

	var exe string

	switch bin {
	case bootstrapper.Player:
		exe = "RobloxPlayerBeta.exe"
	case bootstrapper.Studio:
		exe = "RobloxStudioBeta.exe"
	}

	log.Printf("Launching %s", bin)
	args = append([]string{filepath.Join(verDir, exe)}, args...)

	if err := pfx.Exec(args...); err != nil {
		panic(err)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg := config.Load()
	cfg.Setenv()

	allEnabledDxvk := cfg.Player.Dxvk && cfg.Studio.Dxvk

	pfx := wine.New(dirs.Prefix, "")
	if err := pfx.Setup(); err != nil {
		log.Fatal(err)
	}

	switch os.Args[1] {
	case "player":
		args := os.Args[2:]
		channel := cfg.Player.Channel

		if strings.HasPrefix(strings.Join(args, " "), "roblox-player:1+launchmode:") {
			channel, args = bootstrapper.ParsePlayerURI(args[0])
		}

		if channel != cfg.Player.Channel {
			log.Printf("WARNING: Launch URI has a different channel: %s, forcing user-specified channel", channel)
		}

		Binary(&pfx, bootstrapper.Player, cfg.Player, allEnabledDxvk, args...)
	case "studio":
		Binary(&pfx, bootstrapper.Studio, cfg.Studio, allEnabledDxvk, os.Args[2:]...)
	case "exec":
		if err := pfx.Exec(os.Args[2:]...); err != nil {
			log.Fatal(err)
		}	
	}
}
