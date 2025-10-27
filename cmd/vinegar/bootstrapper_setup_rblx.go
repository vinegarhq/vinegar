package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"sync/atomic"

	"github.com/sewnie/rbxbin"
	"github.com/sewnie/rbxweb"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gtkutil"
	"github.com/vinegarhq/vinegar/internal/netutil"
	"golang.org/x/sync/errgroup"
)

var (
	studio     = rbxweb.BinaryTypeWindowsStudio64
	channelKey = `HKCU\Software\ROBLOX Corporation\Environments\RobloxStudio\Channel`
)

func (b *bootstrapper) setupDeployment() error {
	if err := b.setDeployment(); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	b.dir = filepath.Join(dirs.Versions, b.bin.GUID)

	slog.Info("Using Deployment",
		"guid", b.bin.GUID, "channel", b.bin.Channel)

	if _, err := os.Stat(b.dir); err == nil {
		b.message("Up to date", "guid", b.bin.GUID)
		return nil
	}

	b.message("Installing Studio", "new", b.bin.GUID)
	removeUniqueFiles(dirs.Versions, []string{b.bin.GUID})

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	if err := b.setupPackages(); err != nil {
		return err
	}

	defer b.performing()()

	b.message("Writing AppSettings")
	if err := rbxbin.WriteAppSettings(b.dir); err != nil {
		return fmt.Errorf("appsettings: %w", err)
	}

	// Default channel is none, but UserChannel will set LIVE.
	if b.bin.Channel != "" && b.bin.Channel != "LIVE" {
		b.message("Writing Registry")
		if err := b.pfx.RegistryAdd(channelKey, "www.roblox.com", b.bin.Channel); err != nil {
			return fmt.Errorf("set channel reg: %w", err)
		}
	}

	slog.Info("Successfully installed!", "guid", b.bin.GUID)

	return nil
}

func (b *bootstrapper) setDeployment() error {
	defer b.performing()()

	if b.cfg.Studio.ForcedVersion != "" {
		b.bin = &rbxbin.Deployment{
			Type:    studio,
			Channel: b.cfg.Studio.Channel,
			GUID:    b.cfg.Studio.ForcedVersion,
		}
		return nil
	}

	b.message("Checking for updates")

	d, err := rbxbin.GetDeployment(b.rbx, studio, b.cfg.Studio.Channel)
	if err != nil {
		return err
	}

	gtkutil.IdleAdd(func() {
		b.info.SetLabel(d.Channel)
	})

	b.bin = d

	return nil
}

func (b *bootstrapper) setupPackages() error {
	stop := b.performing()

	b.message("Finding Mirror")
	m, err := rbxbin.GetMirror()
	if err != nil {
		return fmt.Errorf("fetch mirror: %w", err)
	}

	b.message("Fetching Package List")
	pkgs, err := m.GetPackages(b.bin)
	if err != nil {
		return fmt.Errorf("fetch packages: %w", err)
	}

	sums := make([]string, len(pkgs))
	for _, pkg := range pkgs {
		sums = append(sums, pkg.Checksum)
	}
	removeUniqueFiles(dirs.Downloads, sums)

	// Prioritize smaller files first, to have less pressure
	// on network and extraction
	//
	// *Theoretically*, this should be better
	sort.SliceStable(pkgs, func(i, j int) bool {
		return pkgs[i].ZipSize < pkgs[j].ZipSize
	})

	b.message("Fetching Installation Directives")
	pd, err := m.BinaryDirectories(b.bin)
	if err != nil {
		return fmt.Errorf("fetch package dirs: %w", err)
	}

	stop()

	if err := b.installPackages(&m, pkgs, pd); err != nil {
		return err
	}

	return nil
}

func (b *bootstrapper) installPackages(
	mirror *rbxbin.Mirror,
	pkgs []rbxbin.Package,
	pdirs rbxbin.PackageDirectories,
) error {
	total := len(pkgs)
	finished := int64(0)
	group := new(errgroup.Group)

	b.message("Installing Packages", "count", len(pkgs), "dir", b.dir)
	for _, pkg := range pkgs {
		group.Go(func() error {
			if err := b.installPackage(mirror, pdirs, &pkg); err != nil {
				return err
			}

			atomic.AddInt64(&finished, 1)
			gtkutil.IdleAdd(func() {
				b.pbar.SetFraction(float64(finished) / float64(total))
			})

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}

func (b *bootstrapper) installPackage(
	mirror *rbxbin.Mirror,
	pdirs rbxbin.PackageDirectories,
	pkg *rbxbin.Package,
) error {
	src := filepath.Join(dirs.Downloads, pkg.Checksum)
	dst, ok := pdirs[pkg.Name]
	if !ok {
		return fmt.Errorf("unhandled: %s", pkg.Name)
	}

	if err := pkg.Verify(src); err != nil {
		url := mirror.PackageURL(b.bin, pkg.Name)
		slog.Info("Downloading package", "name", pkg.Name, "sum", pkg.Checksum)
		if err := netutil.Download(url, src); err != nil {
			return err
		}
		if err := pkg.Verify(src); err != nil {
			return err
		}
	}

	slog.Info("Extracting package", "name", pkg.Name, "dest", dst)
	return pkg.Extract(src, filepath.Join(b.dir, dst))
}

func removeUniqueFiles(dir string, included []string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		slog.Error("Failed to cleanup directory", "dir", dir, "err", err)
		return
	}

	for _, file := range files {
		if slices.Contains(included, file.Name()) {
			continue
		}

		slog.Info("Removing unique file", "dir", dir, "file", file.Name())
		if err := os.RemoveAll(filepath.Join(dir, file.Name())); err != nil {
			slog.Error("Failed to cleanup file", "err", err)
			break
		}
	}
}
