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
		builder: gtk.NewBuilderFromString(resource("control.ui"), -1),
	}

	ctl.builder.GetObject("window").Cast(&ctl.win)
	ctl.win.SetApplication(&s.app.Application)

	actions := map[string]struct {
		act func() error
		msg string
	}{
		"run-studio":       {nil, "Running Studio"},
		"install-studio":   {nil, "Installing Studio"},
		"uninstall-studio": {ctl.ui.Uninstall, "Uninstalling Studio"},
		"kill-prefix":      {ctl.pfx.Kill, "Stopping Studio"},

		"run-winetricks": {ctl.pfx.Winetricks, "Running Winetricks"},
		"delete-prefix":  {ctl.ui.DeletePrefixes, "Clearing Data"},
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

			stop.SetVisible(false)
			if name == "run-studio" || name == "install-studio" {
				b := ctl.Studio()
				proc := b.Setup
				if name == "run-studio" {
					proc = b.Run
					stop.SetVisible(true)
				}
				action.act = func() error {
					defer Background(b.win.Destroy)
					return proc()
				}
			}

			Background(func() {
				go func() {
					defer ctl.Finished()
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
	var i, u, r, k, d, w gtk.Widget
	ctl.builder.GetObject("install-studio").Cast(&i)
	ctl.builder.GetObject("uninstall-studio").Cast(&u)
	ctl.builder.GetObject("run-studio").Cast(&r)
	ctl.builder.GetObject("kill-prefix").Cast(&k)
	ctl.builder.GetObject("delete-prefix").Cast(&d)
	ctl.builder.GetObject("run-winetricks").Cast(&w)

	empty := dirs.Empty(dirs.Versions)
	i.SetVisible(empty)
	u.SetVisible(!empty)
	r.SetVisible(!empty)

	empty = dirs.Empty(dirs.Prefixes)
	k.SetVisible(!empty)
	d.SetVisible(!empty)
	w.SetVisible(!empty)
}
