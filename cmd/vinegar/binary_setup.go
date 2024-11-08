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
	"github.com/apprehensions/rbxweb/clientsettings"
	cp "github.com/otiai10/copy"
	"github.com/vinegarhq/vinegar/dxvk"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/netutil"
	"golang.org/x/sync/errgroup"
)

func (b *Binary) FetchDeployment() error {
	if b.Config.Channel != "" {
		slog.Warn("Channel is non-default! Only change the deployment channel if you know what you are doing!",
			"channel", b.Config.Channel)
	}

	if b.Config.ForcedVersion != "" {
		slog.Warn("Using forced deployment!", "guid", b.Config.ForcedVersion)

		b.Deploy = rbxbin.Deployment{
			Type:    b.Type,
			Channel: b.Config.Channel,
			GUID:    b.Config.ForcedVersion,
		}
		return nil
	}

	b.Splash.SetMessage("Fetching " + b.Type.Short())

	d, err := rbxbin.GetDeployment(b.Type, b.Config.Channel)
	if err != nil {
		return err
	}

	b.Deploy = d
	return nil
}

func (b *Binary) Setup() error {
	if err := b.FetchDeployment(); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	b.Dir = filepath.Join(dirs.Versions, b.Deploy.GUID)
	b.Splash.SetDesc(fmt.Sprintf("%s %s", b.Deploy.GUID, b.Deploy.Channel))

	if b.State.Version != b.Deploy.GUID {
		slog.Info("Installing Binary", "name", b.Type,
			"old_guid", b.State.Version, "new_guid", b.Deploy.GUID)

		if err := b.Install(); err != nil {
			return fmt.Errorf("install %s: %w", b.Deploy.GUID, err)
		}
	} else {
		slog.Info("Binary is up to date!", "name", b.Type, "guid", b.Deploy.GUID)
	}

	b.Config.Env.Setenv()

	if err := b.SetupOverlay(); err != nil {
		return fmt.Errorf("setup overlay: %w", err)
	}

	if err := b.Config.FFlags.Apply(b.Dir); err != nil {
		return fmt.Errorf("apply fflags: %w", err)
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

func (b *Binary) SetupOverlay() error {
	dir := filepath.Join(dirs.Overlays, strings.ToLower(b.Type.Short()))

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
	b.Splash.SetMessage("Installing " + b.Type.Short())

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	if err := b.SetupPackages(); err != nil {
		return err
	}

	if b.Type == clientsettings.WindowsStudio64 {
		brokenFont := filepath.Join(b.Dir, "StudioFonts", "SourceSansPro-Black.ttf")

		slog.Info("Removing broken font", "path", brokenFont)
		if err := os.RemoveAll(brokenFont); err != nil {
			return err
		}
	}

	if err := rbxbin.WriteAppSettings(b.Dir); err != nil {
		return fmt.Errorf("appsettings: %w", err)
	}

	if err := b.GlobalState.CleanPackages(); err != nil {
		return fmt.Errorf("clean packages: %w", err)
	}

	if err := b.GlobalState.CleanVersions(); err != nil {
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

		if p.Name == "RobloxPlayerLauncher.exe" {
			continue
		}

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

	b.State.Version = b.Deploy.GUID
	for _, pkg := range pkgs {
		b.State.Packages = append(b.State.Packages, pkg.Checksum)
	}
	return nil
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
		slog.Info("DXVK up to date!", "version", b.State.DxvkVersion)
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
