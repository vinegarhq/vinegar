package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync/atomic"

	"github.com/sewnie/rbxbin"
	"github.com/sewnie/rbxweb"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/netutil"
	"golang.org/x/sync/errgroup"
)

var studio = rbxweb.BinaryTypeWindowsStudio64

func (b *bootstrapper) setupDeployment() error {
	if err := b.stepFetchDeployment(); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	b.dir = filepath.Join(dirs.Versions, b.bin.GUID)

	slog.Info("Using Deployment",
		"guid", b.bin.GUID, "channel", b.bin.Channel)

	if b.state.Studio.Version == b.bin.GUID {
		b.message("Up to date", "guid", b.bin.GUID)
		return nil
	}

	b.message("Installing Studio",
		"old", b.state.Studio.Version, "new", b.bin.GUID)

	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	if err := b.stepSetupPackages(); err != nil {
		return err
	}

	if err := b.setupDeploymentFiles(); err != nil {
		return err
	}

	slog.Info("Successfuly installed!", "guid", b.bin.GUID)

	return nil
}

func (b *bootstrapper) stepFetchDeployment() error {
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

	uiThread(func() {
		b.info.SetLabel(d.Channel)
	})

	b.bin = d

	return nil
}

func (b *bootstrapper) stepSetupPackages() error {
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

	// Prioritize smaller files first, to have less pressure
	// on network and extraction
	//
	// *Theoretically*, this should be better
	sort.SliceStable(pkgs, func(i, j int) bool {
		return pkgs[i].ZipSize < pkgs[j].ZipSize
	})

	b.message("Fetching Installation Directives")
	pd, err := m.BinaryDirectories(b.rbx, b.bin)
	if err != nil {
		return fmt.Errorf("fetch package dirs: %w", err)
	}

	stop()

	if err := b.stepPackagesInstall(&m, pkgs, pd); err != nil {
		return err
	}

	return nil
}

func (b *bootstrapper) stepPackagesInstall(
	mirror *rbxbin.Mirror,
	pkgs []rbxbin.Package,
	pdirs rbxbin.PackageDirectories,
) error {
	total := len(pkgs) * 2 // download & extraction
	var finished atomic.Uint32
	group := new(errgroup.Group)

	update := func() {
		finished.Add(1)
		uiThread(func() { b.pbar.SetFraction(float64(finished.Load()) / float64(total)) })
	}

	b.message("Installing Packages", "count", len(pkgs), "dir", b.dir)
	for _, pkg := range pkgs {
		group.Go(func() error {
			return b.stepPackageInstall(mirror, &pkg, pdirs, update)
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	b.state.Studio.Version = b.bin.GUID
	for _, pkg := range pkgs {
		b.state.Studio.Packages = append(b.state.Studio.Packages, pkg.Checksum)
	}

	return nil
}

func (b *bootstrapper) stepPackageInstall(
	mirror *rbxbin.Mirror,
	pkg *rbxbin.Package,
	pdirs rbxbin.PackageDirectories,
	update func(),
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
	update()

	slog.Info("Extracting package", "name", pkg.Name, "dest", dst)
	if err := pkg.Extract(src, filepath.Join(b.dir, dst)); err != nil {
		return err
	}
	update()

	return nil
}

func (b *bootstrapper) setupDeploymentFiles() error {
	defer b.performing()()

	b.message("Writing AppSettings")
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
