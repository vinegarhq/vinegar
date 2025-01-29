package main

import (
	"log/slog"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
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
		return false
	}
	ctl.win.ConnectCloseRequest(&destroy)

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
	}

	ctl.builder.GetObject("stack").Cast(&ctl.stack)

	var label gtk.Label
	var stop gtk.Button
	ctl.builder.GetObject("loading-label").Cast(&label)
	ctl.builder.GetObject("loading-stop").Cast(&stop)

	ctl.UpdateButtons()

	for name, action := range actions {
		act := gio.NewSimpleAction(name, nil)
		actcb := func(_ gio.SimpleAction, _ uintptr) {
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
							s.presentError(err)
						})
					}
				}()
			})
		}
		act.ConnectActivate(&actcb)
		s.app.AddAction(act)
		act.Unref()
	}

	ctl.win.Present()
	ctl.win.Unref()

	return ctl
}

func (ctl *control) UpdateButtons() {
	pfx := dirs.Empty(dirs.Prefixes)
	vers := dirs.Empty(dirs.Versions)

	// While kill-prefix is more of a wineprefix-specific action,
	// it is instead listed as an option belonging to the Studio
	// area, to indicate that it is used to kill studio.
	var inst, uninst, run, kill gtk.Widget
	var del, tricks gtk.Widget
	// init-prefix is always shown
	ctl.builder.GetObject("install-studio").Cast(&inst)
	ctl.builder.GetObject("uninstall-studio").Cast(&uninst)
	ctl.builder.GetObject("run-studio").Cast(&run)
	ctl.builder.GetObject("kill-prefix").Cast(&kill)
	ctl.builder.GetObject("delete-prefix").Cast(&del)
	ctl.builder.GetObject("run-winetricks").Cast(&tricks)
	inst.SetVisible(vers)
	uninst.SetVisible(!vers)
	run.SetVisible(!vers)
	del.SetVisible(!pfx)
	kill.SetVisible(!pfx)
	tricks.SetVisible(!pfx)
}
