package main

import (
	"log"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/config/state"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gui"
	"github.com/vinegarhq/vinegar/roblox"
	"github.com/vinegarhq/vinegar/roblox/bootstrapper"
	"github.com/vinegarhq/vinegar/wine"
	"github.com/vinegarhq/vinegar/wine/dxvk"
)

type Binary struct {
	UI *gui.UI

	GlobalConfig *config.Config
	Config       *config.Application

	Alias   string
	Name    string
	Dir     string
	Prefix  *wine.Prefix
	Type    roblox.BinaryType
	Version roblox.Version
}

func NewBinary(bt roblox.BinaryType, cfg *config.Config, pfx *wine.Prefix) Binary {
	var bcfg config.Application

	switch bt {
	case roblox.Player:
		bcfg = cfg.Player
	case roblox.Studio:
		bcfg = cfg.Studio
	}

	return Binary{
		UI: gui.New(&cfg.UI),

		GlobalConfig: cfg,
		Config:       &bcfg,

		Alias:  bt.String(),
		Name:   bt.BinaryName(),
		Type:   bt,
		Prefix: pfx,
	}
}

func (b *Binary) Run(args ...string) error {
	cmd, err := b.Command(args...)
	if err != nil {
		return err
	}

	log.Printf("Launching %s", b.Name)
	b.UI.Message("Launching " + b.Alias)

	if err := cmd.Start(); err != nil {
		return err
	}

	time.Sleep(1 * time.Second)
	b.UI.Close()
	cmd.Wait()

	if b.Config.AutoKillPrefix {
		b.Prefix.Kill()
	}

	return nil
}

func (b *Binary) FetchVersion() (roblox.Version, error) {
	b.UI.Message("Fetching " + b.Alias)

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

	b.UI.Desc(fmt.Sprintf("%s %s", ver.GUID, ver.Channel))
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

	b.UI.Progress(1.0)
	return nil
}

func (b *Binary) Install() error {
	b.UI.Message("Installing " + b.Alias)

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

	b.UI.Message("Downloading " + b.Alias)
	if err := b.DownloadPackages(&manifest); err != nil {
		return err
	}

	b.UI.Message("Extracting " + b.Alias)
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
		b.UI.Progress(float32(donePkgs) / float32(pkgs))

		return nil
	})
}

func (b *Binary) ExtractPackages(m *bootstrapper.Manifest) error {
	donePkgs := 0
	pkgs := len(m.Packages)
	pkgDirs := bootstrapper.BinaryDirectories(b.Type)

	log.Printf("Extracting %d Packages", pkgs)

	return m.Packages.Perform(func(pkg bootstrapper.Package) error {
		dest, ok := pkgDirs[pkg.Name]

		if !ok {
			return fmt.Errorf("unhandled package: %s", pkg.Name)
		}

		err := pkg.Extract(
			filepath.Join(dirs.Downloads, pkg.Checksum),
			filepath.Join(b.Dir, dest),
		)
		if err != nil {
			return err
		}

		donePkgs++
		b.UI.Progress(float32(donePkgs) / float32(pkgs))

		return nil
	})
}

func (b *Binary) SetupDxvk() error {
	ver, err := state.DxvkVersion()
	if err != nil {
		return err
	}
	installed := ver != ""

	if installed && !b.GlobalConfig.Player.Dxvk && !b.GlobalConfig.Studio.Dxvk {
		b.UI.Message("Uninstalling DXVK")
		if err := dxvk.Remove(b.Prefix); err != nil {
			return err
		}

		return state.SaveDxvk("")
	}

	if !b.Config.Dxvk {
		return nil
	}

	b.UI.Progress(0.0)
	dxvk.Setenv()

	if installed || b.GlobalConfig.DxvkVersion == ver {
		return nil
	}

	if err := dirs.Mkdirs(dirs.Cache); err != nil {
		return err
	}
	path := filepath.Join(dirs.Cache, "dxvk-"+b.GlobalConfig.DxvkVersion+".tar.gz")

	b.UI.Progress(0.3)
	b.UI.Message("Downloading DXVK")
	if err := dxvk.Fetch(path, b.GlobalConfig.DxvkVersion); err != nil {
		return err
	}

	b.UI.Progress(0.7)
	b.UI.Message("Extracting DXVK")
	if err := dxvk.Extract(path, b.Prefix); err != nil {
		return err
	}
	b.UI.Progress(1.0)

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
