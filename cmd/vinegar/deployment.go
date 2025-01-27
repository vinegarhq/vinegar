package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync/atomic"

	"github.com/apprehensions/rbxbin"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/netutil"
	"golang.org/x/sync/errgroup"
)

func (b *bootstrapper) SetupDeployment() error {
	b.Message("Checking for updates")

	if err := b.FetchDeployment(); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	slog.Info("Using Binary Deployment",
		"guid", b.bin.GUID, "channel", b.bin.Channel)

	b.dir = filepath.Join(dirs.Versions, b.bin.GUID)

	if b.state.Studio.Version != b.bin.GUID {
		slog.Info("Studio out of date, installing latest version...",
			"old", b.state.Studio.Version, "new", b.bin.GUID)

		if err := b.Install(); err != nil {
			return fmt.Errorf("install %s: %w", b.bin.GUID, err)
		}
	} else {
		b.status.SetLabel("Up to date")
		slog.Info("Studio up to date!", "guid", b.bin.GUID)
	}

	return nil
}

func (b *bootstrapper) FetchDeployment() error {
	if b.cfg.Studio.ForcedVersion != "" {
		slog.Warn("Using forced deployment!", "guid", b.cfg.Studio.ForcedVersion)

		b.bin = rbxbin.Deployment{
			Type:    Studio,
			Channel: b.cfg.Studio.Channel,
			GUID:    b.cfg.Studio.ForcedVersion,
		}
		return nil
	}

	d, err := rbxbin.GetDeployment(Studio, b.cfg.Studio.Channel)
	if err != nil {
		return err
	}

	b.bin = d
	return nil
}

func (b *bootstrapper) Install() error {
	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	if err := b.SetupPackages(); err != nil {
		return err
	}

	b.Message("Writing AppSettings")
	if err := rbxbin.WriteAppSettings(b.dir); err != nil {
		return fmt.Errorf("appsettings: %w", err)
	}

	brokenFont := filepath.Join(b.dir, "StudioFonts", "SourceSansPro-Black.ttf")
	slog.Info("Removing broken font", "path", brokenFont)
	if err := os.RemoveAll(brokenFont); err != nil {
		return err
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
	b.Message("Finding Mirror")
	m, err := rbxbin.GetMirror()
	if err != nil {
		return fmt.Errorf("fetch mirror: %w", err)
	}

	b.Message("Fetching Package list")
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

	b.Message("Fetching Installation directives")
	pd, err := m.BinaryDirectories(b.bin)
	if err != nil {
		return fmt.Errorf("fetch package dirs: %w", err)
	}

	total := len(pkgs) * 2 // download & extraction
	var done atomic.Uint32
	eg := new(errgroup.Group)

	update := func() {
		done.Add(1)
		b.pbar.SetFraction(float64(done.Load()) / float64(total))
	}

	b.status.SetLabel("Installing")
	slog.Info("Installing Packages", "count", len(pkgs), "dir", b.dir)
	for _, p := range pkgs {
		p := p

		eg.Go(func() error {
			src := filepath.Join(dirs.Downloads, p.Checksum)
			dst, ok := pd[p.Name]
			if !ok {
				return fmt.Errorf("unhandled package: %s", p.Name)
			}

			if err := p.Verify(src); err != nil {
				url := m.PackageURL(b.bin, p.Name)
				slog.Info("Downloading Package", "name", p.Name)

				if err := netutil.Download(url, src); err != nil {
					return err
				}

				slog.Info("Verifying Package", "name", p.Name, "sum", p.Checksum)
				if err := p.Verify(src); err != nil {
					return err
				}
			} else {
				slog.Info("Package cached", "name", p.Name, "sum", p.Checksum)
				update()
			}

			slog.Info("Extracted package", "name", p.Name, "dest", dst)
			if err := p.Extract(src, filepath.Join(b.dir, dst)); err != nil {
				return err
			}
			update()

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

func (ui *ui) Uninstall() error {
	slog.Info("Deleting all deployments!")

	if err := os.RemoveAll(dirs.Versions); err != nil {
		return err
	}

	ui.state.Studio.Version = ""
	ui.state.Studio.Packages = nil

	if err := ui.state.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}
