package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vinegarhq/vinegar/internal/config/state"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/roblox/bootstrapper"
	"github.com/vinegarhq/vinegar/wine"
	"github.com/vinegarhq/vinegar/wine/dxvk"
)

func (b *Binary) FetchVersion() (roblox.Version, error) {
	b.SendLog("Fetching Roblox")

	if b.bcfg.ForcedVersion != "" {
		log.Printf("WARNING: using forced version: %s", b.bcfg.ForcedVersion)

		return roblox.NewVersion(b.btype, b.bcfg.Channel, b.bcfg.ForcedVersion)
	}

	return roblox.LatestVersion(b.btype, b.bcfg.Channel)
}

func (b *Binary) Setup() error {
	ver, err := b.FetchVersion()
	if err != nil {
		return err
	}
	b.ver = ver
	b.dir = filepath.Join(dirs.Versions, ver.GUID)

	stateVer, err := state.Version(b.btype)
	if err != nil {
		log.Printf("Failed to retrieve stored %s version: %s", b.name, err)
	}

	if stateVer != ver.GUID {
		log.Printf("Installing %s (%s -> %s)", b.name, stateVer, ver)

		if err := b.Install(); err != nil {
			return err
		}
	} else {
		log.Printf("%s is up to date (%s)", b.name, ver.GUID)
	}

	b.bcfg.Env.Setenv()

	if err := b.bcfg.FFlags.SetRenderer(b.bcfg.Renderer); err != nil {
		return err
	}

	if err := b.bcfg.FFlags.Apply(b.dir); err != nil {
		return err
	}

	if err := dirs.OverlayDir(b.dir); err != nil {
		return err
	}

	if err := b.SetupDxvk(); err != nil {
		return err
	}

	b.SendProgress(1.0)
	return nil
}

func (b *Binary) Install() error {
	b.SendLog("Installing Roblox")

	manifest, err := bootstrapper.Fetch(b.ver, dirs.Downloads)
	if err != nil {
		return err
	}

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	b.SendLog("Downloading Roblox")
	if err := b.DownloadPackages(&manifest); err != nil {
		return err
	}

	b.SendLog("Extracting Roblox")
	if err := b.ExtractPackages(&manifest); err != nil {
		return err
	}

	if err := bootstrapper.WriteAppSettings(b.dir); err != nil {
		return err
	}

	if err := state.SaveManifest(&manifest); err != nil {
		return err
	}

	if err := state.CleanPackages(); err != nil {
		return err
	}

	return state.CleanVersions()
}

func (b *Binary) DownloadPackages(m *bootstrapper.Manifest) error {
	donePkgs := 0
	pkgs := len(m.Packages)

	log.Printf("Downloading %d Packages", pkgs)

	return m.Packages.Perform(func(pkg bootstrapper.Package) error {
		err := pkg.Fetch(filepath.Join(dirs.Downloads, pkg.Checksum), m.DeployURL)
		if err != nil {
			return err
		}

		donePkgs++
		b.SendProgress(float32(donePkgs) / float32(pkgs))

		return nil
	})
}

func (b *Binary) ExtractPackages(m *bootstrapper.Manifest) error {
	donePkgs := 0
	pkgs := len(m.Packages)
	pkgDirs := bootstrapper.Directories(b.btype)

	log.Printf("Extracting %d Packages", pkgs)

	return m.Packages.Perform(func(pkg bootstrapper.Package) error {
		dest, ok := pkgDirs[pkg.Name]

		if !ok {
			return fmt.Errorf("unhandled package: %s", pkg.Name)
		}

		err := pkg.Extract(
			filepath.Join(dirs.Downloads, pkg.Checksum),
			filepath.Join(b.dir, dest),
		)
		if err != nil {
			return err
		}

		donePkgs++
		b.SendProgress(float32(donePkgs) / float32(pkgs))

		return nil
	})
}

func (b *Binary) SetupDxvk() error {
	ver, err := state.DxvkVersion()
	if err != nil {
		return err
	}
	installed := ver != ""

	if installed && !b.cfg.Player.Dxvk && !b.cfg.Studio.Dxvk {
		b.SendLog("Uninstalling DXVK")
		if err := dxvk.Remove(b.pfx); err != nil {
			return err
		}

		return state.SaveDxvk("")
	}

	if !b.bcfg.Dxvk {
		return nil
	}

	b.SendProgress(0.0)
	dxvk.Setenv()

	if installed || b.cfg.DxvkVersion == ver {
		return nil
	}

	if err := dirs.Mkdirs(dirs.Cache); err != nil {
		return err
	}
	path := filepath.Join(dirs.Cache, "dxvk-"+b.cfg.DxvkVersion+".tar.gz")

	b.SendProgress(0.3)
	b.SendLog("Downloading DXVK")
	if err := dxvk.Fetch(path, b.cfg.DxvkVersion); err != nil {
		return err
	}

	b.SendProgress(0.7)
	b.SendLog("Extracting DXVK")
	if err := dxvk.Extract(path, b.pfx); err != nil {
		return err
	}
	b.SendProgress(1.0)

	return state.SaveDxvk(b.cfg.DxvkVersion)
}

func (b *Binary) Command(args ...string) (*wine.Cmd, error) {
	if strings.HasPrefix(strings.Join(args, " "), "roblox-studio:1") {
		args = []string{"-protocolString", args[0]}
	}

	if b.cfg.MultipleInstances {
		mutexer := b.pfx.Command("wine", filepath.Join(BinPrefix, "robloxmutexer.exe"))
		err := mutexer.Start()
		if err != nil {
			return &wine.Cmd{}, err
		}
	}

	cmd := b.pfx.Wine(filepath.Join(b.dir, b.btype.Executable()), args...)

	launcher := strings.Fields(b.bcfg.Launcher)
	if len(launcher) >= 1 {
		cmd.Args = append(launcher, cmd.Args...)

		launcherPath, err := exec.LookPath(launcher[0])
		if err != nil {
			return &wine.Cmd{}, err
		}

		cmd.Path = launcherPath
	}

	return cmd, nil
}
