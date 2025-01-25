package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/apprehensions/rbxbin"
	cp "github.com/otiai10/copy"
	"github.com/vinegarhq/vinegar/dxvk"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/netutil"
	"golang.org/x/sync/errgroup"
)

func (b *Binary) FetchDeployment() error {
	if b.Config.Studio.Channel != "" {
		slog.Warn("Channel is non-default! Only change the deployment channel if you know what you are doing!",
			"channel", b.Config.Studio.Channel)
	}

	if b.Config.Studio.ForcedVersion != "" {
		slog.Warn("Using forced deployment!", "guid", b.Config.Studio.ForcedVersion)

		b.Deploy = rbxbin.Deployment{
			Type:    Studio,
			Channel: b.Config.Studio.Channel,
			GUID:    b.Config.Studio.ForcedVersion,
		}
		return nil
	}

	b.Splash.SetMessage("Fetching " + Studio.Short())

	d, err := rbxbin.GetDeployment(Studio, b.Config.Studio.Channel)
	if err != nil {
		return err
	}

	b.Deploy = d
	return nil
}

func (b *Binary) Setup() error {
	// Player is no longer supported by Vinegar, remove unnecessary data
	if b.State.Player.Version != "" || b.State.Player.DxvkVersion != "" {
		os.RemoveAll(filepath.Join(dirs.Versions, b.State.Player.Version))
		os.RemoveAll(filepath.Join(dirs.Prefixes, "player"))
		b.State.Player.DxvkVersion = ""
		b.State.Player.Version = ""
		b.State.Player.Packages = nil
	}

	if err := b.FetchDeployment(); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	b.Dir = filepath.Join(dirs.Versions, b.Deploy.GUID)
	b.Splash.SetDesc(fmt.Sprintf("%s %s", b.Deploy.GUID, b.Deploy.Channel))

	if b.State.Studio.Version != b.Deploy.GUID {
		slog.Info("Installing Binary", "name", Studio,
			"old_guid", b.State.Studio.Version, "new_guid", b.Deploy.GUID)

		if err := b.Install(); err != nil {
			return fmt.Errorf("install %s: %w", b.Deploy.GUID, err)
		}
	} else {
		slog.Info("Binary is up to date!", "name", Studio, "guid", b.Deploy.GUID)
	}

	b.Config.Studio.Env.Setenv()

	if err := b.SetupOverlay(); err != nil {
		return fmt.Errorf("setup overlay: %w", err)
	}

	if err := b.Config.Studio.FFlags.Apply(b.Dir); err != nil {
		return fmt.Errorf("apply fflags: %w", err)
	}

	if err := b.SetupDxvk(); err != nil {
		return fmt.Errorf("setup dxvk %s: %w", b.Config.Studio.DxvkVersion, err)
	}

	b.Splash.SetProgress(1.0)
	if err := b.State.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

func (b *Binary) SetupOverlay() error {
	dir := filepath.Join(dirs.Overlays, strings.ToLower(Studio.Short()))

	// Don't copy Overlay if it doesn't exist
	_, err := os.Stat(dir)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	slog.Info("Copying Overlay directory's files", "src", dir, "path", b.Dir)
	b.Splash.SetMessage("Copying Overlay")

	return cp.Copy(dir, b.Dir)
}

func (b *Binary) Install() error {
	b.Splash.SetMessage("Installing " + Studio.Short())

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	if err := b.SetupPackages(); err != nil {
		return err
	}

	brokenFont := filepath.Join(b.Dir, "StudioFonts", "SourceSansPro-Black.ttf")

	slog.Info("Removing broken font", "path", brokenFont)
	if err := os.RemoveAll(brokenFont); err != nil {
		return err
	}

	if err := rbxbin.WriteAppSettings(b.Dir); err != nil {
		return fmt.Errorf("appsettings: %w", err)
	}

	if err := b.State.CleanPackages(); err != nil {
		return fmt.Errorf("clean packages: %w", err)
	}

	if err := b.State.CleanVersions(); err != nil {
		return fmt.Errorf("clean versions: %w", err)
	}

	return nil
}

func (b *Binary) SetupPackages() error {
	m, err := rbxbin.GetMirror()
	if err != nil {
		return fmt.Errorf("fetch mirror: %w", err)
	}

	pkgs, err := m.GetPackages(b.Deploy)
	if err != nil {
		return fmt.Errorf("fetch packages: %w", err)
	}

	// Prioritize smaller files first, to have less pressure
	// on network and extraction
	//
	// *Theoretically*, this should be better
	sort.SliceStable(pkgs, func(i, j int) bool {
		return pkgs[i].ZipSize < pkgs[j].ZipSize
	})

	slog.Info("Fetching package directories")

	pd, err := m.BinaryDirectories(b.Deploy)
	if err != nil {
		return fmt.Errorf("fetch package dirs: %w", err)
	}

	total := len(pkgs) * 2 // download & extraction
	var done atomic.Uint32
	eg := new(errgroup.Group)

	slog.Info("Installing Packages", "count", len(pkgs))
	for _, p := range pkgs {
		p := p

		eg.Go(func() error {
			src := filepath.Join(dirs.Downloads, p.Checksum)
			dst, ok := pd[p.Name]
			if !ok {
				return fmt.Errorf("unhandled package: %s", p.Name)
			}

			defer func() {
				done.Add(1)
				b.Splash.SetProgress(float32(done.Load()) / float32(total))
			}()

			if err := p.Verify(src); err != nil {
				url := m.PackageURL(b.Deploy, p.Name)
				slog.Info("Downloading Package", "name", p.Name)

				if err := netutil.Download(url, src); err != nil {
					return err
				}

				if err := p.Verify(src); err != nil {
					return err
				}
			} else {
				slog.Info("Package is already downloaded", "name", p.Name)
			}

			if err := p.Extract(src, filepath.Join(b.Dir, dst)); err != nil {
				return err
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	b.State.Studio.Version = b.Deploy.GUID
	for _, pkg := range pkgs {
		b.State.Studio.Packages = append(b.State.Studio.Packages, pkg.Checksum)
	}
	return nil
}

func (b *Binary) SetupDxvk() error {
	if b.State.Studio.DxvkVersion != "" && !b.Config.Studio.Dxvk {
		b.Splash.SetMessage("Uninstalling DXVK")
		if err := dxvk.Remove(b.Prefix); err != nil {
			return fmt.Errorf("remove dxvk: %w", err)
		}

		b.State.Studio.DxvkVersion = ""
		return nil
	}

	if !b.Config.Studio.Dxvk {
		return nil
	}

	b.Splash.SetProgress(0.0)
	dxvk.Setenv()

	if b.Config.Studio.DxvkVersion == b.State.Studio.DxvkVersion {
		slog.Info("DXVK up to date!", "version", b.State.Studio.DxvkVersion)
		return nil
	}

	dxvkPath := filepath.Join(dirs.Cache, "dxvk-"+b.Config.Studio.DxvkVersion+".tar.gz")

	if _, err := os.Stat(dxvkPath); err != nil {
		url := dxvk.URL(b.Config.Studio.DxvkVersion)

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

	b.State.Studio.DxvkVersion = b.Config.Studio.DxvkVersion
	return nil
}
