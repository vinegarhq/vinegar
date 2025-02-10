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
	"github.com/apprehensions/wine/dxvk"
	"github.com/apprehensions/wine/webview"
	cp "github.com/otiai10/copy"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/netutil"
	"golang.org/x/sync/errgroup"
)

var studio = clientsettings.WindowsStudio64

func (b *bootstrapper) prepare() error {
	b.message("Applying Environment")
	dxvk.Setenv(b.cfg.Studio.Dxvk)

	// Required to read Roblox logs.
	b.cfg.Studio.Env["WINEDEBUG"] += ",warn+debugstr"
	b.cfg.Studio.Env.Setenv()

	if err := b.setupOverlay(); err != nil {
		return fmt.Errorf("setup overlay: %w", err)
	}

	b.message("Applying FFlags")
	if err := b.cfg.Studio.FFlags.Apply(b.dir); err != nil {
		return fmt.Errorf("apply fflags: %w", err)
	}

	return nil
}

func (b *bootstrapper) setup() error {
	b.removePlayer()

	if err := b.setupPrefix(); err != nil {
		return fmt.Errorf("prefix: %w", err)
	}

	if err := b.setupDeployment(); err != nil {
		return err
	}

	if err := b.setupDxvk(); err != nil {
		return fmt.Errorf("setup dxvk %s: %w", b.cfg.Studio.DxvkVersion, err)
	}

	if err := b.state.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

func (b *bootstrapper) setupPrefix() error {
	b.message("Setting up Wine")

	if c := b.pfx.Wine(""); c.Err != nil {
		return fmt.Errorf("wine: %w", c.Err)
	}

	if b.pfx.Exists() {
		return nil
	}

	stop := b.performing()

	b.message("Initializing Wineprefix", "dir", b.pfx.Dir())
	if err := run(b.pfx.Init()); err != nil {
		return err
	}

	stop()

	if err := b.webViewInstall(); err != nil {
		return err
	}

	return nil
}

func (b *bootstrapper) webViewInstall() error {
	if b.cfg.Studio.WebView == "" {
		return nil
	}

	name := filepath.Join(dirs.Cache, "webview-"+b.cfg.Studio.WebView+".exe")
	if _, err := os.Stat(name); err != nil {
		if err := b.webViewDownload(name); err != nil {
			return err
		}
	}

	defer b.performing()()

	b.message("Setting Wineprefix DPI")
	// Studio will not load past the splash screen if the DPI
	// is 96 and WebView is installed.
	if err := b.pfx.SetDPI(98); err != nil {
		return nil
	}

	b.message("Setting Wineprefix version")
	// If it is not win7, WebView will appear black.
	if err := run(b.pfx.Wine("winecfg", "/v", "win7")); err != nil {
		return nil
	}

	b.message("Installing WebView", "path", name)

	return run(webview.Install(b.pfx, name))
}

func (b *bootstrapper) webViewDownload(name string) error {
	stop := b.performing()

	b.message("Fetching WebView", "upload", b.cfg.Studio.WebView)
	d, err := webview.GetDownload(b.cfg.Studio.WebView)
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp("", "unc_msedgestandalone.*.exe")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if err != nil {
		return err
	}

	stop()
	b.message("Downloading WebView", "version", d.Version)
	err = netutil.DownloadProgress(d.URL, tmp.Name(), &b.pbar)
	if err != nil {
		return err
	}

	cac, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer cac.Close()

	return d.Extract(tmp, cac)
}

func (b *bootstrapper) setupOverlay() error {
	dir := filepath.Join(dirs.Overlays, strings.ToLower(studio.Short()))

	// Don't copy Overlay if it doesn't exist
	_, err := os.Stat(dir)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	b.message("Copying Overlay")

	return cp.Copy(dir, b.dir)
}

func (b *bootstrapper) setupDeployment() error {
	b.message("Checking for updates")

	if err := b.fetchDeployment(); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	b.dir = filepath.Join(dirs.Versions, b.bin.GUID)

	if b.state.Studio.Version != b.bin.GUID {
		slog.Info("Studio out of date, installing latest version...",
			"old", b.state.Studio.Version, "new", b.bin.GUID)

		if err := b.install(); err != nil {
			return fmt.Errorf("install %s: %w", b.bin.GUID, err)
		}
	} else {
		b.message("Up to date", "guid", b.bin.GUID)
	}

	return nil
}

func (b *bootstrapper) fetchDeployment() error {
	defer b.performing()()

	if b.cfg.Studio.ForcedVersion != "" {
		b.bin = rbxbin.Deployment{
			Type:    studio,
			Channel: b.cfg.Studio.Channel,
			GUID:    b.cfg.Studio.ForcedVersion,
		}

		slog.Warn("Using forced deployment!",
			"guid", b.bin.GUID, "channel", b.bin.Channel)
		return nil
	}

	d, err := rbxbin.GetDeployment(studio, b.cfg.Studio.Channel)
	if err != nil {
		return err
	}

	b.bin = d
	slog.Info("Using Binary Deployment",
		"guid", b.bin.GUID, "channel", b.bin.Channel)
	return nil
}

func (b *bootstrapper) install() error {
	if err := dirs.Mkdirs(dirs.Downloads); err != nil {
		return err
	}

	if err := b.setupPackages(); err != nil {
		return err
	}

	if err := b.setupInstall(); err != nil {
		return err
	}

	// MIME has to be setup in-order for Studio to login.
	if err := b.setupMIME(); err != nil {
		return err
	}

	slog.Info("Successfuly installed!", "guid", b.bin.GUID)

	return nil
}

func (b *bootstrapper) setupPackages() error {
	stop := b.performing()

	b.message("Finding Mirror")
	m, err := rbxbin.GetMirror()
	if err != nil {
		return fmt.Errorf("fetch mirror: %w", err)
	}

	b.message("Fetching Package List", "channel", b.bin.Channel)
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
	pd, err := m.BinaryDirectories(b.bin)
	if err != nil {
		return fmt.Errorf("fetch package dirs: %w", err)
	}

	total := len(pkgs) * 2 // download & extraction
	var done atomic.Uint32
	eg := new(errgroup.Group)

	update := func() {
		done.Add(1)
		idle(func() { b.pbar.SetFraction(float64(done.Load()) / float64(total)) })
	}

	stop()
	b.message("Installing Studio", "count", len(pkgs), "dir", b.dir)
	for _, p := range pkgs {
		eg.Go(func() error {
			return b.setupPackage(pd, &m, &p, update)
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

func (b *bootstrapper) setupPackage(
	pd rbxbin.PackageDirectories,
	m *rbxbin.Mirror,
	p *rbxbin.Package,
	update func(),
) error {
	src := filepath.Join(dirs.Downloads, p.Checksum)
	dst, ok := pd[p.Name]
	if !ok {
		return fmt.Errorf("unhandled: %s", p.Name)
	}

	if err := p.Verify(src); err != nil {
		url := m.PackageURL(b.bin, p.Name)
		slog.Info("Downloading package", "name", p.Name, "sum", p.Checksum)
		if err := netutil.Download(url, src); err != nil {
			return err
		}
		if err := p.Verify(src); err != nil {
			return err
		}
	}
	update()

	slog.Info("Extracting package", "name", p.Name, "dest", dst)
	if err := p.Extract(src, filepath.Join(b.dir, dst)); err != nil {
		return err
	}
	update()

	return nil
}

func (b *bootstrapper) setupInstall() error {
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

func (b *bootstrapper) setupDxvk() error {
	if !b.cfg.Studio.Dxvk ||
		b.cfg.Studio.DxvkVersion == b.state.Studio.DxvkVersion {
		return nil
	}

	ver := b.cfg.Studio.DxvkVersion
	dxvkPath := filepath.Join(dirs.Cache, "dxvk-"+ver+".tar.gz")

	if err := dirs.Mkdirs(dirs.Cache); err != nil {
		return err
	}

	if _, err := os.Stat(dxvkPath); err != nil {
		url := dxvk.URL(ver)

		b.message("Downloading DXVK", "ver", ver)

		if err := netutil.DownloadProgress(url, dxvkPath, &b.pbar); err != nil {
			return fmt.Errorf("download: %w", err)
		}
	}

	defer b.performing()()

	b.message("Extracting DXVK", "version", ver)

	if err := dxvk.Extract(b.pfx, dxvkPath); err != nil {
		return err
	}

	b.state.Studio.DxvkVersion = ver

	return nil
}
