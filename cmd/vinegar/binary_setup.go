package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/state"
	"github.com/vinegarhq/vinegar/roblox"
	boot "github.com/vinegarhq/vinegar/roblox/bootstrapper"
	"github.com/vinegarhq/vinegar/wine/dxvk"
	"golang.org/x/sync/errgroup"
)

func (b *Binary) FetchDeployment() error {
	b.Splash.SetMessage("Fetching " + b.Alias)

	if b.Config.ForcedVersion != "" {
		log.Printf("WARNING: using forced version: %s", b.Config.ForcedVersion)

		d := boot.NewDeployment(b.Type, b.Config.Channel, b.Config.ForcedVersion)
		b.Deploy = &d
		return nil
	}

	d, err := boot.FetchDeployment(b.Type, b.Config.Channel)
	if err != nil {
		return err
	}

	b.Deploy = &d
	return nil
}

func (b *Binary) Setup() error {
	s, err := state.Load()
	if err != nil {
		return err
	}
	b.State = &s

	if err := b.FetchDeployment(); err != nil {
		return err
	}

	b.Dir = filepath.Join(dirs.Versions, b.Deploy.GUID)
	b.Splash.SetDesc(fmt.Sprintf("%s %s", b.Deploy.GUID, b.Deploy.Channel))

	stateVer := b.State.Version(b.Type)
	if stateVer != b.Deploy.GUID {
		log.Printf("Installing %s (%s -> %s)", b.Name, stateVer, b.Deploy.GUID)

		if err := b.Install(); err != nil {
			return err
		}
	} else {
		log.Printf("%s is up to date (%s)", b.Name, b.Deploy.GUID)
	}

	b.Config.Env.Setenv()

	log.Println("Using Renderer:", b.Config.Renderer)
	if err := b.Config.FFlags.Apply(b.Dir); err != nil {
		return err
	}

	if err := dirs.OverlayDir(b.Dir); err != nil {
		return err
	}

	if err := b.SetupDxvk(); err != nil {
		return err
	}

	b.Splash.SetProgress(1.0)
	return b.State.Save()
}

func (b *Binary) Install() error {
	b.Splash.SetMessage("Installing " + b.Alias)

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	pm, err := boot.FetchPackageManifest(b.Deploy)
	if err != nil {
		return err
	}

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	// Prioritize smaller files first, to have less pressure
	// on network and extraction
	//
	// *Theoretically*, this should be better
	sort.SliceStable(pm.Packages, func(i, j int) bool {
		return pm.Packages[i].ZipSize < pm.Packages[j].ZipSize
	})

	b.Splash.SetMessage("Downloading " + b.Alias)
	if err := b.DownloadPackages(&pm); err != nil {
		return err
	}

	b.Splash.SetMessage("Extracting " + b.Alias)
	if err := b.ExtractPackages(&pm); err != nil {
		return err
	}

	if b.Type == roblox.Studio {
		brokenFont := filepath.Join(b.Dir, "StudioFonts", "SourceSansPro-Black.ttf")

		log.Printf("Removing broken font %s", brokenFont)
		if err := os.RemoveAll(brokenFont); err != nil {
			log.Printf("Failed to remove font: %s", err)
		}
	}

	if err := boot.WriteAppSettings(b.Dir); err != nil {
		return err
	}

	b.State.AddBinary(&pm)

	if err := b.State.CleanPackages(); err != nil {
		return err
	}

	return b.State.CleanVersions()
}

func (b *Binary) PerformPackages(pm *boot.PackageManifest, fn func(boot.Package) error) error {
	donePkgs := 0
	pkgsLen := len(pm.Packages)
	eg := new(errgroup.Group)

	for _, p := range pm.Packages {
		p := p
		eg.Go(func() error {
			err := fn(p)
			if err != nil {
				return err
			}

			donePkgs++
			b.Splash.SetProgress(float32(donePkgs) / float32(pkgsLen))

			return nil
		})
	}

	return eg.Wait()
}

func (b *Binary) DownloadPackages(pm *boot.PackageManifest) error {
	log.Printf("Downloading %d Packages for %s", len(pm.Packages), pm.Deployment.GUID)

	return b.PerformPackages(pm, func(pkg boot.Package) error {
		return pkg.Download(filepath.Join(dirs.Downloads, pkg.Checksum), pm.DeployURL)
	})
}

func (b *Binary) ExtractPackages(pm *boot.PackageManifest) error {
	log.Printf("Extracting %d Packages for %s", len(pm.Packages), pm.Deployment.GUID)

	pkgDirs := boot.BinaryDirectories(b.Type)

	return b.PerformPackages(pm, func(pkg boot.Package) error {
		dest, ok := pkgDirs[pkg.Name]

		if !ok {
			return fmt.Errorf("unhandled package: %s", pkg.Name)
		}

		return pkg.Extract(filepath.Join(dirs.Downloads, pkg.Checksum), filepath.Join(b.Dir, dest))
	})
}

func (b *Binary) SetupDxvk() error {
	if b.State.DxvkVersion != "" && !b.GlobalConfig.Player.Dxvk && !b.GlobalConfig.Studio.Dxvk {
		b.Splash.SetMessage("Uninstalling DXVK")
		if err := dxvk.Remove(b.Prefix); err != nil {
			return err
		}

		b.State.DxvkVersion = ""
		return nil
	}

	if !b.Config.Dxvk {
		return nil
	}

	b.Splash.SetProgress(0.0)
	dxvk.Setenv()

	if b.GlobalConfig.DxvkVersion == b.State.DxvkVersion {
		return nil
	}

	// This would only get saved if Install succeeded
	b.State.DxvkVersion = b.GlobalConfig.DxvkVersion

	b.Splash.SetMessage("Installing DXVK")
	return dxvk.Install(b.GlobalConfig.DxvkVersion, b.Prefix)
}
