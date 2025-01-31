package main

import (
	"errors"
	"log/slog"
	"os"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	_ "github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/vinegarhq/vinegar/internal/dirs"
)

type control struct {
	*ui

	builder *gtk.Builder
	win     adw.ApplicationWindow

	boot *bootstrapper

	stack adw.ViewStack
}

func (ctl *control) Finished() {
	Background(func() {
		ctl.UpdateButtons()
		ctl.stack.SetVisibleChildName("control")
	})
}

func (ctl *control) Studio() *bootstrapper {
	b := ctl.NewBootstrapper()

	destroy := func(_ gtk.Window) bool {
		// BUG: realistically no other way to cancel
		//      the bootstrapper, so just exit immediately!
		b.app.Quit()
		return false
	}
	b.win.ConnectCloseRequest(&destroy)

	return b
}

func (s *ui) NewControl() control {
	ctl := control{
		ui:      s,
		boot:    s.NewBootstrapper(),
		builder: gtk.NewBuilderFromString(resource("control.ui"), -1),
	}

	ctl.builder.GetObject("window").Cast(&ctl.win)
	ctl.win.SetApplication(&s.app.Application)
	destroy := func(_ gtk.Window) bool {
		s.app.Quit()
		return false
	}
	ctl.win.ConnectCloseRequest(&destroy)

	// For the time being, use in-house editing.
	// ctl.SetupConfigurationActions()
	ctl.SetupControlActions()

	ctl.win.Present()
	ctl.win.Unref()

	return ctl
}

func (ctl *control) SetupControlActions() {
	actions := map[string]struct {
		act  func() error
		msg  string
		boot bool
	}{
		"run-studio":       {ctl.boot.Run, "Running Studio", true},
		"install-studio":   {ctl.boot.Setup, "Installing Studio", true},
		"uninstall-studio": {ctl.DeleteDeployments, "Uninstalling Studio", false},
		"kill-prefix":      {ctl.KillPrefix, "Stopping Studio", false},
		"init-prefix":      {ctl.boot.PrefixInit, "Initializing Data", true},
		"delete-prefix":    {ctl.DeletePrefixes, "Clearing Data", false},
		"run-winetricks":   {ctl.RunWinetricks, "Running Winetricks", false},
		"clear-cache":      {ctl.CacheClear, "Clearing Downloads", false},

		"save-config": {ctl.SaveConfig, "Loading Configuration", false},
	}

	ctl.builder.GetObject("stack").Cast(&ctl.stack)
	var label gtk.Label
	var stop gtk.Button
	ctl.builder.GetObject("loading-label").Cast(&label)
	ctl.builder.GetObject("loading-stop").Cast(&stop)

	ctl.PutConfig()
	ctl.UpdateButtons()

	for name, action := range actions {
		act := gio.NewSimpleAction(name, nil)
		actcb := func(_ gio.SimpleAction, p uintptr) {

			ctl.stack.SetVisibleChildName("loading")
			label.SetLabel(action.msg + "...")
			stop.SetVisible(name == "run-studio")

			if action.boot {
				Background(func() {
					ctl.win.SetTransientFor(&ctl.boot.win.Window)
					ctl.boot.win.Show()
				})
			}
			Background(func() {
				go func() {
					defer ctl.Finished()
					if action.boot {
						defer Background(ctl.boot.win.Hide)
					}
					if err := action.act(); err != nil {
						slog.Error("Error occurred while running action",
							"action", name, "err", err)
						Background(func() {
							ctl.presentSimpleError(err)
						})
					}
				}()
			})
		}
		act.ConnectActivate(&actcb)
		ctl.app.AddAction(act)
		act.Unref()
	}
}

/*
func (ctl *control) SetupConfigurationActions() {
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

func (ctl *control) PutConfig() {
	var view gtk.TextView
	ctl.builder.GetObject("config-view").Cast(&view)

	b, err := os.ReadFile(dirs.ConfigPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		b = []byte(err.Error())
	}

	view.GetBuffer().SetText(string(b), -1)
}

func (ctl *control) SaveConfig() error {
	var view gtk.TextView
	var start, end gtk.TextIter
	ctl.builder.GetObject("config-view").Cast(&view)

	buf := view.GetBuffer()
	buf.GetBounds(&start, &end)
	text := buf.GetText(&start, &end, false)

	slog.Info("Saving Configuration!")
	if err := os.WriteFile(dirs.ConfigPath, []byte(text), 0664); err != nil {
		return err
	}

	return ctl.LoadConfig()
}

func (ctl *control) UpdateButtons() {
	pfx := dirs.Empty(dirs.Prefixes)
	vers := dirs.Empty(dirs.Versions)

	// While kill-prefix is more of a wineprefix-specific action,
	// it is instead listed as an option belonging to the Studio
	// area, to indicate that it is used to kill studio.
	var inst, uninst, run, kill gtk.Widget
	var del, tricks gtk.Widget

	ctl.builder.GetObject("studio-install").Cast(&inst)
	ctl.builder.GetObject("studio-uninstall").Cast(&uninst)
	ctl.builder.GetObject("studio-run").Cast(&run)
	// prefix-init is always shown
	ctl.builder.GetObject("prefix-kill").Cast(&kill)
	ctl.builder.GetObject("prefix-delete").Cast(&del)
	ctl.builder.GetObject("prefix-winetricks").Cast(&tricks)

	inst.SetVisible(vers)
	uninst.SetVisible(!vers)
	run.SetVisible(!vers)
	del.SetVisible(!pfx)
	kill.SetVisible(!pfx)
	tricks.SetVisible(!pfx)
}
