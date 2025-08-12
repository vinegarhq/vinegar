package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/vinegarhq/vinegar/internal/dirs"
)

type control struct {
	*app

	builder *gtk.Builder
	win     adw.ApplicationWindow
}

func (s *app) newControl() control {
	ctl := control{
		app:     s,
		builder: gtk.NewBuilderFromResource("/org/vinegarhq/Vinegar/ui/control.ui"),
	}

	ctl.builder.GetObject("window").Cast(&ctl.win)
	ctl.win.SetApplication(&s.Application.Application)

	abt := gio.NewSimpleAction("about", nil)
	abtcb := func(_ gio.SimpleAction, p uintptr) {
		w := adw.NewAboutDialogFromAppdata("/org/vinegarhq/Vinegar/metainfo.xml", version[1:])
		w.Present(&ctl.win.Widget)
		w.SetDebugInfo(s.debugInfo())
		w.Unref()
	}
	abt.ConnectActivate(&abtcb)
	ctl.AddAction(abt)
	abt.Unref()

	ctl.configPut()
	ctl.updateButtons()

	// For the time being, use in-house editing.
	// ctl.setupConfigurationActions()
	ctl.setupControlActions()

	ctl.win.Present()
	ctl.win.Unref()

	return ctl
}

func (ctl *control) configPut() {
	var view gtk.TextView
	ctl.builder.GetObject("config-view").Cast(&view)

	b, err := os.ReadFile(dirs.ConfigPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		b = []byte(err.Error())
	}

	view.GetBuffer().SetText(string(b), -1)
}

func (ctl *control) setupControlActions() {
	actions := map[string]struct {
		msg string
		act interface{}
	}{
		"run-studio": {"Executing Studio", (*bootstrapper).start},
		"run-winecfg": {"Running Wine Configurator", func() error {
			return run(ctl.pfx.Wine("winecfg"))
		}},

		"install-studio":   {"Installing Studio", (*bootstrapper).setup},
		"uninstall-studio": {"Deleting all deployments", ctl.deleteDeployments},

		"init-prefix":   {"Initializing Wineprefix", (*bootstrapper).setupPrefix},
		"kill-prefix":   {"Killing Wineprefix", ctl.pfx.Kill},
		"delete-prefix": {"Deleting Wineprefix", ctl.deletePrefixes},

		"save-config": {"Saving configuration to file", ctl.configSave},
		"clear-cache": {"Cleaning up cache folder", ctl.clearCache},
	}

	var stack adw.ViewStack
	ctl.builder.GetObject("stack").Cast(&stack)
	var label gtk.Label
	ctl.builder.GetObject("loading-label").Cast(&label)

	// Reserved only for execution of studio
	var btn gtk.Button
	ctl.builder.GetObject("loading-stop").Cast(&btn)
	stop := gio.NewSimpleAction("show-stop", nil)
	stopcb := func(_ gio.SimpleAction, p uintptr) {
		btn.SetVisible(true)
	}
	stop.ConnectActivate(&stopcb)
	ctl.AddAction(stop)
	stop.Unref()

	for name, action := range actions {
		act := gio.NewSimpleAction(name, nil)
		actcb := func(_ gio.SimpleAction, p uintptr) {
			stack.SetVisibleChildName("loading")
			label.SetLabel(action.msg + "...")
			btn.SetVisible(false)

			var run func() error
			switch f := action.act.(type) {
			case func() error:
				run = func() error {
					slog.Info(action.msg + "...")
					return f()
				}
			case func(*bootstrapper) error:
				b := ctl.newBootstrapper()
				b.win.SetTransientFor(&ctl.win.Window)
				run = func() error {
					defer idle(b.win.Destroy)
					return f(b)
				}
			default:
				panic("unreachable")
			}

			var tf glib.ThreadFunc = func(uintptr) uintptr {
				defer idle(func() {
					ctl.updateButtons()
					stack.SetVisibleChildName("control")
				})
				if err := run(); err != nil {
					idle(func() { ctl.error(err) })
				}
				return 0
			}
			glib.NewThread("action", &tf, 0)
		}

		act.ConnectActivate(&actcb)
		ctl.AddAction(act)
		act.Unref()
	}
}

/*
func (ctl *control) setupConfigurationActions() {
	props := []struct {
		name   string
		widget string
		modify any
	}{
		{"gamemode", "switch", ctl.cfg.Studio.GameMode},
		{"launcher", "entry", ctl.cfg.Studio.Launcher},
	}

	save := func() {
		if err := ctl.LoadConfig(); err != nil {
			ctl.presentSimpleError(err)
		}
	}

	for _, p := range props {
		obj := ctl.builder.GetObject(p.name)

		switch p.widget {
		case "switch":
			// AdwSwitchRow does not implement Gtk.Switch, and neither
			// does puregotk have a reference for AdwSwitchRow.
			actcb := func() {
				var new bool
				obj.Get("active", &new)
				p.modify = new
				slog.Info("Configuration Switch", "name", p.widget, "value", p.modify)
				save()
			}
			var w gtk.Widget
			obj.Cast(&w)
			obj.ConnectSignal("notify::active", &actcb)
			slog.Info("Setup Widget", "name", p.name)
		case "entry":
			var entry adw.EntryRow
			obj.Cast(&entry)

			cb := func(_ adw.EntryRow) {
				p.modify = entry.GetText()
				slog.Info("Configuration Entry", "name", p.widget, "value", p.modify)
				save()
			}
			entry.ConnectApply(&cb)
		default:
			panic("unreachable")
		}
	}

	// this is a special button!
	var rootRow adw.ActionRow
	ctl.builder.GetObject("wineroot-row").Cast(&rootRow)
	rootRow.SetSubtitle(ctl.cfg.Studio.WineRoot)

	var rootSelect gtk.Button
	ctl.builder.GetObject("wineroot-select").Cast(&rootSelect)
	actcb := func(_ gtk.Button) {
		// gtk.FileChooser is deprecated for gtk.FileDialog, but puregotk does not have it.
		fc := gtk.NewFileChooserDialog("Select Wine Installation", &ctl.win.Window,
			gtk.FileChooserActionSelectFolderValue,
			"Cancel", gtk.ResponseCancelValue,
			"Select", gtk.ResponseAcceptValue,
			unsafe.Pointer(nil),
		)
		rcb := func(_ gtk.Dialog, _ int) {
			cp := ctl.cfg.Studio
			cp.WineRoot = fc.GetFile().GetPath()
			// if err := cp.Setup(); err != nil {
			rootRow.SetSubtitle(ctl.cfg.Studio.WineRoot)
			fc.Destroy()
		}
		fc.ConnectResponse(&rcb)
		fc.Present()
	}
	rootSelect.ConnectClicked(&actcb)
}
*/

func (ctl *control) configSave() error {
	var view gtk.TextView
	var start, end gtk.TextIter
	ctl.builder.GetObject("config-view").Cast(&view)

	buf := view.GetBuffer()
	buf.GetBounds(&start, &end)
	text := buf.GetText(&start, &end, false)

	if err := dirs.Mkdirs(dirs.Config); err != nil {
		return err
	}

	f, err := os.OpenFile(dirs.ConfigPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write([]byte(text)); err != nil {
		return err
	}

	return ctl.loadConfig()
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
		if path == dirs.Cache || path == dirs.Logs || path == ctl.logFile.Name() {
			return nil
		}

		slog.Info("Removing cache file", "path", path)
		return os.RemoveAll(path)
	})
}

func (ctl *control) updateButtons() {
	pfx := ctl.pfx.Exists()
	vers := dirs.Empty(dirs.Versions)

	// While kill-prefix is more of a wineprefix-specific action,
	// it is instead listed as an option belonging to the Studio
	// area, to indicate that it is used to kill studio.
	var inst, uninst, run gtk.Widget
	ctl.builder.GetObject("studio-install").Cast(&inst)
	ctl.builder.GetObject("studio-uninstall").Cast(&uninst)
	ctl.builder.GetObject("studio-run").Cast(&run)
	inst.SetVisible(vers)
	uninst.SetVisible(!vers)
	run.SetVisible(!vers)

	var init, kill, del, cfg gtk.Widget
	ctl.builder.GetObject("prefix-init").Cast(&init)
	ctl.builder.GetObject("prefix-kill").Cast(&kill)
	ctl.builder.GetObject("prefix-delete").Cast(&del)
	ctl.builder.GetObject("prefix-configure").Cast(&cfg)
	init.SetVisible(!pfx)
	del.SetVisible(pfx)
	kill.SetVisible(pfx)
	cfg.SetVisible(pfx)
}
