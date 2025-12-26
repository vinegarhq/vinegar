package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gutil"
	"github.com/vinegarhq/vinegar/internal/logging"
)

func (m *manager) runWineCmd() error {
	cmd := m.runner.GetText()
	args := strings.Fields(cmd)
	if len(args) < 1 {
		return nil
	}
	if err := m.startWine(); err != nil {
		return err
	}
	slog.Info("Running Wine command", "args", args)
	return m.pfx.Wine(args[0], args[1:]...).Run()
}

func (m *manager) hideRunUntil() func() {
	var button gtk.Button
	var stack gtk.Stack
	m.builder.GetObject("stack").Cast(&stack)
	m.builder.GetObject("btn-run").Cast(&button)
	gutil.IdleAdd(func() {
		stack.SetVisibleChildName("stkpage-spinner")
		button.SetSensitive(false)
	})
	return func() {
		gutil.IdleAdd(func() {
			button.SetSensitive(true)
			stack.SetVisibleChildName("stkpage-btn")
			m.updateRun()
		})
	}
}

func (m *manager) run() error {
	show := m.hideRunUntil()
	defer show()

	// "Run Studio"
	if m.pfx.Exists() && len(m.boot.procs) == 0 {
		visible := func() {
			if !m.boot.win.GetVisible() {
				show()
			}
		}
		h := m.app.boot.win.ConnectSignal("notify::visible", &visible)
		defer func() { m.app.boot.win.DisconnectSignal(h) }()
		return m.app.boot.run()
	}

	// "Stop"
	if len(m.boot.procs) > 0 {
		for _, p := range m.boot.procs {
			slog.Info("Killing Studio", "pid", p.Pid)
			p.Kill()
		}
		m.boot.procs = nil
		return nil
	}

	// "Initialize"
	if err := m.startWine(); err != nil {
		return err
	}
	return nil
}

// No error will be returned, error handling is displayed
// from an AdwBanner.
func (m *manager) saveConfig() {
	var banner adw.Banner
	m.builder.GetObject("banner-config-error").Cast(&banner)
	banner.SetRevealed(false)

	err := m.reload()
	if err != nil {
		banner.SetTitle(fmt.Sprintf("Invalid: %v", err))
		banner.SetRevealed(true)
		slog.Error("Configuration validation error", "err", err)
		return
	}
	slog.Info("Saving configuration!")
	if err := m.cfg.Save(); err != nil {
		m.showError(fmt.Errorf("save config: %w", err))
		return
	}

	u, err := m.pfx.NeedsUpdate()
	if err != nil {
		m.showError(fmt.Errorf("determine update: %w", err))
	} else if u {
		m.errThread(m.startWine)
	}
}

func (m *manager) showAbout() {
	b := gutil.ResourceData(gutil.Resource("metainfo.xml"))
	data := struct {
		XMLName  xml.Name `xml:"component"`
		Releases struct {
			Release []struct {
				Text    string `xml:",chardata"`
				Version string `xml:"version,attr"`
				Date    string `xml:"date,attr"`
			} `xml:"release"`
		} `xml:"releases"`
	}{}

	if err := xml.Unmarshal(b, &data); err != nil {
		panic("expected valid appstream: " + err.Error())
	}

	dialog := adw.NewAboutDialogFromAppdata(gutil.Resource("metainfo.xml"),
		data.Releases.Release[0].Version)
	dialog.Present(&m.win.Widget)
	dialog.Unref()
}

func (m *manager) killPrefix() error {
	if err := m.pfx.Kill(); err != nil {
		return err
	}

	m.showToast("Stopped all processes")
	return nil
}

func (m *manager) deletePrefixes() error {
	defer m.hideRunUntil()()

	_ = m.pfx.Kill()
	if err := os.RemoveAll(dirs.Prefixes); err != nil {
		return err
	}

	m.showToast("Deleted all data")
	m.wine.SetSensitive(false)
	return nil
}

func (m *manager) deleteDeployments() error {
	defer m.hideRunUntil()()

	if err := os.RemoveAll(dirs.Versions); err != nil {
		return err
	}

	m.showToast("Uninstalled studio")
	return nil
}

func (m *manager) clearCache() error {
	err := filepath.WalkDir(dirs.Cache, func(path string, d fs.DirEntry, err error) error {
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
	if err != nil {
		return err
	}

	m.showToast("Cleared cache")
	return nil
}
