package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gtkutil"
	"github.com/vinegarhq/vinegar/internal/logging"
)

func (m *manager) hideRunUntil() func() {
	var button gtk.Button
	var stack gtk.Stack
	m.builder.GetObject("stack").Cast(&stack)
	m.builder.GetObject("btn-run").Cast(&button)
	gtkutil.IdleAdd(func() {
		stack.SetVisibleChildName("stkpage-spinner")
		button.SetSensitive(false)
	})
	return func() {
		gtkutil.IdleAdd(func() {
			button.SetSensitive(true)
			stack.SetVisibleChildName("stkpage-btn")
			m.updateRun()
		})
	}
}

func (m *manager) run() error {
	show := m.hideRunUntil()
	defer show()

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

	if len(m.boot.procs) > 0 {
		return m.pfx.Kill()
	}
	return m.app.boot.setupPrefix()
}

// No error will be returned, error handling is displayed
// from an AdwBanner.
func (m *manager) saveConfig() {
	var banner adw.Banner
	m.builder.GetObject("banner-config-error").Cast(&banner)
	banner.SetRevealed(false)

	if err := m.cfg.Save(); err != nil {
		m.showError(err)
		return
	}

	err := m.reload()
	if err == nil {
		return
	}

	banner.SetTitle(fmt.Sprintf("Invalid: %v", err))
	banner.SetRevealed(true)
	slog.Info("Configuration validation error", "err", err)
	return
}

func (m *manager) showAbout() {
	b := gtkutil.ResourceData(gtkutil.Resource("metainfo.xml"))
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

	dialog := adw.NewAboutDialogFromAppdata(gtkutil.Resource("metainfo.xml"),
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
