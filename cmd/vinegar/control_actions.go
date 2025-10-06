package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gtkutil"
	"github.com/vinegarhq/vinegar/internal/logging"
	"github.com/vinegarhq/vinegar/sysinfo"
)

func (ctl *control) run() error {
	var err error
	if ctl.pfx.Exists() {
		err = ctl.app.boot.run()
	} else {
		err = ctl.app.boot.setupPrefix()
	}
	ctl.updateRun()

	return err
}

// No error will be returned, error handling is displayed
// from an AdwBanner.
func (ctl *control) saveConfig() {
	var banner adw.Banner
	ctl.builder.GetObject("banner-config-error").Cast(&banner)
	banner.SetRevealed(false)

	if err := ctl.cfg.Save(); err != nil {
		ctl.showError(err)
		return
	}

	err := ctl.loadConfig()
	if err == nil {
		return
	}

	banner.SetTitle(fmt.Sprintf("Invalid: %v", err))
	banner.SetRevealed(true)
	slog.Info("Configuration validation error", "err", err)
	return
}

func (ctl *control) showAbout() {
	w := adw.NewAboutDialogFromAppdata(gtkutil.Resource("metainfo.xml"), version[1:])
	w.Present(&ctl.win.Widget)

	var revision string
	bi, _ := debug.ReadBuildInfo()
	for _, bs := range bi.Settings {
		if bs.Key == "vcs.revision" {
			revision = fmt.Sprintf("(%s)", bs.Value)
		}
	}

	var b strings.Builder

	inst := "source"
	if sysinfo.InFlatpak {
		inst = "flatpak"
	}

	info := `* Vinegar: %s %s
* Distro: %s
* Processor: %s
* Kernel: %s
* Wine: %s
* Installation: %s
`

	fmt.Fprintf(&b, info,
		version, revision,
		sysinfo.Distro,
		sysinfo.CPU.Name,
		sysinfo.Kernel,
		ctl.pfx.Version(),
		inst,
	)

	fmt.Fprintln(&b, "* Cards:")
	for i, c := range sysinfo.Cards {
		fmt.Fprintf(&b, "  * Card %d: %s %s %s\n",
			i, c.Driver, filepath.Base(c.Device), c.Path)
	}

	w.SetDebugInfo(b.String())
	w.Unref()
}

func (ctl *control) deleteDeployments() error {
	if err := os.RemoveAll(dirs.Versions); err != nil {
		return err
	}

	ctl.state.Studio.Version = ""
	ctl.state.Studio.Packages = nil

	return ctl.state.Save()
}

func (ctl *control) deletePrefixes() error {
	slog.Info("Deleting Wineprefixes!")
	defer ctl.updateRun()

	if err := ctl.pfx.Kill(); err != nil {
		return fmt.Errorf("kill prefix: %w", err)
	}

	if err := os.RemoveAll(dirs.Prefixes); err != nil {
		return err
	}

	ctl.state.Studio.DxvkVersion = ""

	if err := ctl.state.Save(); err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	return nil
}

func (ctl *control) clearCache() error {
	return filepath.WalkDir(dirs.Cache, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if path == dirs.Cache || path == dirs.Logs || path == logging.Path {
			return nil
		}

		slog.Info("Removing cache file", "path", path)
		return os.RemoveAll(path)
	})
}
