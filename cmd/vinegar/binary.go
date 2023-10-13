package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/config/state"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/splash"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/roblox/bootstrapper"
	"github.com/vinegarhq/vinegar/util"
	"github.com/vinegarhq/vinegar/wine"
	"github.com/vinegarhq/vinegar/wine/dxvk"
)

type Binary struct {
	Splash *splash.Splash

	GlobalConfig *config.Config
	Config       *config.Binary

	Alias   string
	Name    string
	Dir     string
	Prefix  *wine.Prefix
	Type    roblox.BinaryType
	Version roblox.Version
}

func NewBinary(bt roblox.BinaryType, cfg *config.Config, pfx *wine.Prefix) Binary {
	var bcfg config.Binary

	switch bt {
	case roblox.Player:
		bcfg = cfg.Player
	case roblox.Studio:
		bcfg = cfg.Studio
	}

	return Binary{
		Splash: splash.New(&cfg.Splash),

		GlobalConfig: cfg,
		Config:       &bcfg,

		Alias:  bt.String(),
		Name:   bt.BinaryName(),
		Type:   bt,
		Prefix: pfx,
	}
}

func (b *Binary) Run(args ...string) error {
	exe := b.Type.Executable()
	cmd, err := b.Command(args...)
	if err != nil {
		return err
	}

	log.Printf("Launching %s", b.Name)
	b.Splash.Message("Launching " + b.Alias)

	kill := true

	// If roblox is already running, don't kill wineprefix, even if
	// auto kill prefix is enabled
	if util.CommFound("Roblox") {
		log.Println("Roblox is already running, not killing wineprefix after exit")
		kill = false
	}

	// Launches into foreground
	if err := cmd.Start(); err != nil {
		return err
	}

	time.Sleep(2500 * time.Millisecond)
	b.Splash.Close()

	if kill && b.Config.AutoKillPrefix {
		log.Println("Waiting for Roblox's process to die :)")

		for {
			time.Sleep(1 * time.Second)

			if !util.CommFound(exe[:15]) {
				break
			}
		}

		b.Prefix.Kill()
	}

	return nil
}

func (b *Binary) FetchVersion() (roblox.Version, error) {
	b.Splash.Message("Fetching " + b.Alias)

	if b.Config.ForcedVersion != "" {
		log.Printf("WARNING: using forced version: %s", b.Config.ForcedVersion)

		return roblox.NewVersion(b.Type, b.Config.Channel, b.Config.ForcedVersion)
	}

	return roblox.LatestVersion(b.Type, b.Config.Channel)
}

func (b *Binary) Setup() error {
	ver, err := b.FetchVersion()
	if err != nil {
		return err
	}

	b.Splash.Desc(fmt.Sprintf("%s %s", ver.GUID, ver.Channel))
	b.Version = ver
	b.Dir = filepath.Join(dirs.Versions, ver.GUID)

	stateVer, err := state.Version(b.Type)
	if err != nil {
		log.Printf("Failed to retrieve stored %s version: %s", b.Name, err)
	}

	if stateVer != ver.GUID {
		log.Printf("Installing %s (%s -> %s)", b.Name, stateVer, ver)

		if err := b.Install(); err != nil {
			return err
		}
	} else {
		log.Printf("%s is up to date (%s)", b.Name, ver.GUID)
	}

	b.Config.Env.Setenv()

	if err := b.Config.FFlags.SetRenderer(b.Config.Renderer); err != nil {
		return err
	}

	if err := b.Config.FFlags.Apply(b.Dir); err != nil {
		return err
	}

	if err := dirs.OverlayDir(b.Dir); err != nil {
		return err
	}

	if err := b.SetupDxvk(); err != nil {
		return err
	}

	b.Splash.Progress(1.0)
	return nil
}

func (b *Binary) Install() error {
	b.Splash.Message("Installing " + b.Alias)

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	manifest, err := bootstrapper.FetchManifest(&b.Version)
	if err != nil {
		return err
	}

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	b.Splash.Message("Downloading " + b.Alias)
	if err := b.DownloadPackages(&manifest); err != nil {
		return err
	}

	b.Splash.Message("Extracting " + b.Alias)
	if err := b.ExtractPackages(&manifest); err != nil {
		return err
	}

	if err := bootstrapper.WriteAppSettings(b.Dir); err != nil {
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

func (b *Binary) PerformPackages(m *bootstrapper.Manifest, fn func(bootstrapper.Package) error) error {
	donePkgs := 0
	pkgsLen := len(m.Packages)

	return m.Packages.Perform(func(pkg bootstrapper.Package) error {
		err := fn(pkg)
		if err != nil {
			return err
		}

		donePkgs++
		b.Splash.Progress(float32(donePkgs) / float32(pkgsLen))

		return nil
	})
}

func (b *Binary) DownloadPackages(m *bootstrapper.Manifest) error {
	log.Printf("Downloading %d Packages for %s", len(m.Packages), m.Version.GUID)

	return b.PerformPackages(m, func(pkg bootstrapper.Package) error {
		return pkg.Fetch(filepath.Join(dirs.Downloads, pkg.Checksum), m.DeployURL)
	})
}

func (b *Binary) ExtractPackages(m *bootstrapper.Manifest) error {
	log.Printf("Extracting %d Packages for %s", len(m.Packages), m.Version.GUID)
	pkgDirs := bootstrapper.BinaryDirectories(b.Type)

	return b.PerformPackages(m, func(pkg bootstrapper.Package) error {
		dest, ok := pkgDirs[pkg.Name]

		if !ok {
			return fmt.Errorf("unhandled package: %s", pkg.Name)
		}

		return pkg.Extract(filepath.Join(dirs.Downloads, pkg.Checksum), filepath.Join(b.Dir, dest))
	})
}

func (b *Binary) SetupDxvk() error {
	ver, err := state.DxvkVersion()
	if err != nil {
		return err
	}
	installed := ver != ""

	if installed && !b.GlobalConfig.Player.Dxvk && !b.GlobalConfig.Studio.Dxvk {
		b.Splash.Message("Uninstalling DXVK")
		if err := dxvk.Remove(b.Prefix); err != nil {
			return err
		}

		return state.SaveDxvk("")
	}

	if !b.Config.Dxvk {
		return nil
	}

	b.Splash.Progress(0.0)
	dxvk.Setenv()

	if b.GlobalConfig.DxvkVersion == ver {
		return nil
	}

	if err := dirs.Mkdirs(dirs.Cache); err != nil {
		return err
	}
	path := filepath.Join(dirs.Cache, "dxvk-"+b.GlobalConfig.DxvkVersion+".tar.gz")

	b.Splash.Progress(0.3)
	b.Splash.Message("Downloading DXVK")
	if err := dxvk.Fetch(path, b.GlobalConfig.DxvkVersion); err != nil {
		return err
	}

	b.Splash.Progress(0.7)
	b.Splash.Message("Extracting DXVK")
	if err := dxvk.Extract(path, b.Prefix); err != nil {
		return err
	}
	b.Splash.Progress(1.0)

	return state.SaveDxvk(b.GlobalConfig.DxvkVersion)
}

func (b *Binary) Command(args ...string) (*wine.Cmd, error) {
	if strings.HasPrefix(strings.Join(args, " "), "roblox-studio:1") {
		args = []string{"-protocolString", args[0]}
	}

	if b.GlobalConfig.MultipleInstances {
		mutexer := b.Prefix.Command("wine", filepath.Join(BinPrefix, "robloxmutexer.exe"))
		err := mutexer.Start()
		if err != nil {
			return &wine.Cmd{}, err
		}
	}

	cmd := b.Prefix.Wine(filepath.Join(b.Dir, b.Type.Executable()), args...)

	launcher := strings.Fields(b.Config.Launcher)
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
