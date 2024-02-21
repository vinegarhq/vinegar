package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cp "github.com/otiai10/copy"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/netutil"
	"github.com/vinegarhq/vinegar/roblox"
	boot "github.com/vinegarhq/vinegar/roblox/bootstrapper"
	"github.com/vinegarhq/vinegar/wine/dxvk"
	"golang.org/x/sync/errgroup"
)

func (b *Binary) SetDeployment() error {
	if b.Config.Channel != "" {
		slog.Warn("Channel is non-default! Only change the deployment channel if you know what you are doing!",
			"channel", b.Config.Channel)
	}

	if b.Config.ForcedVersion != "" {
		slog.Warn("Using forced deployment!", "guid", b.Config.ForcedVersion)

		d := boot.NewDeployment(b.Type, b.Config.Channel, b.Config.ForcedVersion)
		b.Deploy = &d
		return nil
	}

	b.Splash.SetMessage("Fetching " + b.Alias)

	d, err := boot.FetchDeployment(b.Type, b.Config.Channel)
	if err != nil {
		return err
	}

	b.Deploy = &d
	return nil
}

func (b *Binary) Setup() error {
	if err := b.SetDeployment(); err != nil {
		return fmt.Errorf("set %s deployment: %w", b.Config.Channel, err)
	}

	b.Dir = filepath.Join(dirs.Versions, b.Deploy.GUID)
	b.Splash.SetDesc(fmt.Sprintf("%s %s", b.Deploy.GUID, b.Deploy.Channel))

	if b.State.Version != b.Deploy.GUID {
		slog.Info("Installing Binary", "name", b.Name,
			"old_guid", b.State.Version, "new_guid", b.Deploy.GUID)

		if err := b.Install(); err != nil {
			return fmt.Errorf("install %s: %w", b.Deploy.GUID, err)
		}
	} else {
		slog.Info("Binary is up to date!", "name", b.Name, "guid", b.Deploy.GUID)
	}

	b.Config.Env.Setenv()

	if err := b.Config.FFlags.Apply(b.Dir); err != nil {
		return fmt.Errorf("apply fflags: %w", err)
	}

	overlayDir := filepath.Join(dirs.Overlays, strings.ToLower(b.Type.String()))
	_, err := os.Stat(overlayDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat overlay: %w", err)
	} else if err == nil {
		slog.Info("Copying Overlay directory's files", "src", overlayDir, "path", b.Dir)

		if err := cp.Copy(overlayDir, b.Dir); err != nil {
			return fmt.Errorf("overlay dir: %w", err)
		}
	}

	if err := b.SetupDxvk(); err != nil {
		return fmt.Errorf("setup dxvk %s: %w", b.Config.DxvkVersion, err)
	}

	b.Splash.SetProgress(1.0)
	if err := b.GlobalState.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

func (b *Binary) Install() error {
	b.Splash.SetMessage("Installing " + b.Alias)

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	pm, err := boot.FetchPackageManifest(b.Deploy)
	if err != nil {
		return fmt.Errorf("fetch package manifest: %w", err)
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
		return fmt.Errorf("download: %w", err)
	}

	b.Splash.SetMessage("Extracting " + b.Alias)
	if err := b.ExtractPackages(&pm); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	if b.Type == roblox.Studio {
		brokenFont := filepath.Join(b.Dir, "StudioFonts", "SourceSansPro-Black.ttf")

		slog.Info("Removing broken font", "path", brokenFont)
		if err := os.RemoveAll(brokenFont); err != nil {
			return err
		}
	}

	if err := boot.WriteAppSettings(b.Dir); err != nil {
		return fmt.Errorf("appsettings: %w", err)
	}

	b.State.Add(&pm)

	if err := b.GlobalState.CleanPackages(); err != nil {
		return fmt.Errorf("clean packages: %w", err)
	}

	if err := b.GlobalState.CleanVersions(); err != nil {
		return fmt.Errorf("clean versions: %w", err)
	}

	return nil
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
	slog.Info("Downloading Packages", "guid", pm.Deployment.GUID, "count", len(pm.Packages))

	return b.PerformPackages(pm, func(pkg boot.Package) error {
		return pkg.Download(filepath.Join(dirs.Downloads, pkg.Checksum), pm.DeployURL)
	})
}

func (b *Binary) ExtractPackages(pm *boot.PackageManifest) error {
	slog.Info("Extracting Packages", "guid", pm.Deployment.GUID, "count", len(pm.Packages))

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
	if b.State.DxvkVersion != "" && !b.Config.Dxvk {
		b.Splash.SetMessage("Uninstalling DXVK")
		if err := dxvk.Remove(b.Prefix); err != nil {
			return fmt.Errorf("remove dxvk: %w", err)
		}

		b.State.DxvkVersion = ""
		return nil
	}

	if !b.Config.Dxvk {
		return nil
	}

	b.Splash.SetProgress(0.0)
	dxvk.Setenv()

	if b.Config.DxvkVersion == b.State.DxvkVersion {
		return nil
	}

	dxvkPath := filepath.Join(dirs.Cache, "dxvk-"+b.Config.DxvkVersion+".tar.gz")

	if _, err := os.Stat(dxvkPath); err != nil {
		url := dxvk.URL(b.Config.DxvkVersion)

		b.Splash.SetMessage("Downloading DXVK")
		slog.Info("Downloading DXVK tarball", "url", url, "path", dxvkPath)

		if err := netutil.DownloadProgress(url, dxvkPath, b.Splash.SetProgress); err != nil {
			return fmt.Errorf("download: %w", err)
		}
	}

	b.Splash.SetProgress(1.0)
	b.Splash.SetMessage("Installing DXVK")

	if err := dxvk.Extract(dxvkPath, b.Prefix); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	b.State.DxvkVersion = b.Config.DxvkVersion
	return nil
}
