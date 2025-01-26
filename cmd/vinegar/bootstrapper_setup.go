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

func (b *bootstrapper) FetchDeployment() error {
	if b.cfg.Studio.Channel != "" {
		slog.Warn("Channel is non-default! Only change the deployment channel if you know what you are doing!",
			"channel", b.cfg.Studio.Channel)
	}

	if b.cfg.Studio.ForcedVersion != "" {
		slog.Warn("Using forced deployment!", "guid", b.cfg.Studio.ForcedVersion)

		b.bin = rbxbin.Deployment{
			Type:    Studio,
			Channel: b.cfg.Studio.Channel,
			GUID:    b.cfg.Studio.ForcedVersion,
		}
		return nil
	}

	b.status.SetLabel("Fetching")

	d, err := rbxbin.GetDeployment(Studio, b.cfg.Studio.Channel)
	if err != nil {
		return err
	}

	b.bin = d
	return nil
}

func (b *bootstrapper) Setup() error {
	// Player is no longer supported by Vinegar, remove unnecessary data
	if b.state.Player.Version != "" || b.state.Player.DxvkVersion != "" {
		os.RemoveAll(filepath.Join(dirs.Versions, b.state.Player.Version))
		os.RemoveAll(filepath.Join(dirs.Prefixes, "player"))
		b.state.Player.DxvkVersion = ""
		b.state.Player.Version = ""
		b.state.Player.Packages = nil
	}

	if err := b.FetchDeployment(); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	b.dir = filepath.Join(dirs.Versions, b.bin.GUID)

	if b.state.Studio.Version != b.bin.GUID {
		slog.Info("Installing bootstrapper", "name", Studio,
			"old_guid", b.state.Studio.Version, "new_guid", b.bin.GUID)

		if err := b.Install(); err != nil {
			return fmt.Errorf("install %s: %w", b.bin.GUID, err)
		}
	} else {
		slog.Info("bootstrapper is up to date!", "name", Studio, "guid", b.bin.GUID)
	}

	b.cfg.Studio.Env.Setenv()

	if err := b.SetupOverlay(); err != nil {
		return fmt.Errorf("setup overlay: %w", err)
	}

	if err := b.cfg.Studio.FFlags.Apply(b.dir); err != nil {
		return fmt.Errorf("apply fflags: %w", err)
	}

	if err := b.SetupDxvk(); err != nil {
		return fmt.Errorf("setup dxvk %s: %w", b.cfg.Studio.DxvkVersion, err)
	}

	b.pbar.SetFraction(1.0)
	if err := b.state.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

func (b *bootstrapper) SetupOverlay() error {
	dir := filepath.Join(dirs.Overlays, strings.ToLower(Studio.Short()))

	// Don't copy Overlay if it doesn't exist
	_, err := os.Stat(dir)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	slog.Info("Copying Overlay directory's files", "src", dir, "path", b.dir)
	b.status.SetLabel("Copying Overlay")

	return cp.Copy(dir, b.dir)
}

func (b *bootstrapper) Install() error {
	b.status.SetLabel("Installing")

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	if err := b.SetupPackages(); err != nil {
		return err
	}

	brokenFont := filepath.Join(b.dir, "StudioFonts", "SourceSansPro-Black.ttf")

	slog.Info("Removing broken font", "path", brokenFont)
	if err := os.RemoveAll(brokenFont); err != nil {
		return err
	}

	if err := rbxbin.WriteAppSettings(b.dir); err != nil {
		return fmt.Errorf("appsettings: %w", err)
	}

	if err := b.state.CleanPackages(); err != nil {
		return fmt.Errorf("clean packages: %w", err)
	}

	if err := b.state.CleanVersions(); err != nil {
		return fmt.Errorf("clean versions: %w", err)
	}

	return nil
}

func (b *bootstrapper) SetupPackages() error {
	m, err := rbxbin.GetMirror()
	if err != nil {
		return fmt.Errorf("fetch mirror: %w", err)
	}

	pkgs, err := m.GetPackages(b.bin)
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

	pd, err := m.BinaryDirectories(b.bin)
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
				b.pbar.SetFraction(float64(done.Load()) / float64(total))
			}()

			if err := p.Verify(src); err != nil {
				url := m.PackageURL(b.bin, p.Name)
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

			if err := p.Extract(src, filepath.Join(b.dir, dst)); err != nil {
				return err
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	b.state.Studio.Version = b.bin.GUID
	for _, pkg := range pkgs {
		b.state.Studio.Packages = append(b.state.Studio.Packages, pkg.Checksum)
	}
	return nil
}

func (b *bootstrapper) SetupDxvk() error {
	if b.state.Studio.DxvkVersion != "" && !b.cfg.Studio.Dxvk {
		b.status.SetLabel("Uninstalling DXVK")
		if err := dxvk.Remove(b.pfx); err != nil {
			return fmt.Errorf("remove dxvk: %w", err)
		}

		b.state.Studio.DxvkVersion = ""
		return nil
	}

	if !b.cfg.Studio.Dxvk {
		return nil
	}

	b.pbar.SetFraction(0.0)
	dxvk.Setenv()

	if b.cfg.Studio.DxvkVersion == b.state.Studio.DxvkVersion {
		slog.Info("DXVK up to date!", "version", b.state.Studio.DxvkVersion)
		return nil
	}

	dxvkPath := filepath.Join(dirs.Cache, "dxvk-"+b.cfg.Studio.DxvkVersion+".tar.gz")

	if _, err := os.Stat(dxvkPath); err != nil {
		url := dxvk.URL(b.cfg.Studio.DxvkVersion)

		b.status.SetLabel("Downloading DXVK")
		slog.Info("Downloading DXVK tarball", "url", url, "path", dxvkPath)

		if err := netutil.DownloadProgress(url, dxvkPath, nil); err != nil {
			return fmt.Errorf("download: %w", err)
		}
	}

	b.pbar.SetFraction(1.0)
	b.status.SetLabel("Installing DXVK")

	if err := dxvk.Extract(dxvkPath, b.pfx); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	b.state.Studio.DxvkVersion = b.cfg.Studio.DxvkVersion
	return nil
}
