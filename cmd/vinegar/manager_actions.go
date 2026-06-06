package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"codeberg.org/puregotk/puregotk/v4/adw"
	"codeberg.org/puregotk/puregotk/v4/gtk"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gutil"
	"github.com/vinegarhq/vinegar/internal/logging"

	. "github.com/pojntfx/go-gettext/pkg/i18n"
)

func (m *manager) runWineCmd(e gtk.Entry) {
	args := strings.Fields(e.GetText())
	if len(args) < 1 {
		return
	}
	m.errThread(func() error {
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
	dialog := adw.NewAboutDialogFromAppdata(gutil.Resource("metainfo.xml"), m.version)
	dialog.Present(&m.win.Widget)
	dialog.Unref()
}

func (m *manager) killPrefix() error {
	if err := m.pfx.Kill(); err != nil {
		return err
	}

	m.showToast(L("Stopped all processes"))
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

func (m *manager) deleteDeployments() error {
	if err := os.RemoveAll(dirs.Versions); err != nil {
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
