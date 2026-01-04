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

	. "github.com/pojntfx/go-gettext/pkg/i18n"
)

func (m *manager) run() error {
	stop := m.loading()

	// "Run Studio"
	if m.pfx.Exists() && len(m.boot.procs) == 0 {
		visible := func() {
			if !m.boot.win.GetVisible() {
				stop()
			}
		}
		// Bootstrapper automatically hides itself after running Studio,
		// hook on it to signify to the manager
		h := m.app.boot.win.ConnectSignal("notify::visible", &visible)
		defer func() {
			m.app.boot.win.DisconnectSignal(h)
			m.updateRunContent()
		}()
		return m.app.boot.run()
	}

	defer stop()

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
	_, err := m.prepareWine()
	return err
}

func (m *manager) runWineCmd(e gtk.Entry) {
	stop := m.loading()
	args := strings.Fields(e.GetText())
	if len(args) < 1 {
		return
	}
	m.errThread(func() error {
		defer stop()
		if _, err := m.prepareWine(); err != nil {
			return err
		}
		return m.pfx.Wine(args[0], args[1:]...).Run()
	})
}

func (m *manager) saveConfig() {
	m.applyConfig()

	slog.Info("Saving configuration!")
	if err := m.cfg.Save(); err != nil {
		// in this case, it would probably be a TOML or
		// I/O error.
		m.showError(fmt.Errorf("save config: %w", err))
		return
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

	m.showToast(L("Stopped all processes"))
	slog.Debug("Processes", "boot", m.boot.procs, "pfx", m.pfx.Running())
	return nil
}

func (m *manager) deletePrefixes() error {
	_ = m.pfx.Kill()
	if err := os.RemoveAll(dirs.Prefixes); err != nil {
		return err
	}

	m.showToast(L("Deleted Wine data"))
	return nil
}

func (m *manager) deleteDeployment() error {
	if err := os.RemoveAll(dirs.Deployment); err != nil {
		return err
	}

	m.showToast(L("Uninstalled studio"))
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

	m.showToast(L("Cleared cache"))
	return nil
}
