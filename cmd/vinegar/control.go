package main

import (
	"log/slog"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

type control struct {
	*ui

	builder *gtk.Builder
	win     adw.ApplicationWindow

	stack adw.ViewStack
}

func (ctl *control) Finished() {
	ctl.stack.SetVisibleChildName("control")
}

func (ctl *control) Studio() bootstrapper {
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
		"run-studio":       {nil, "Running Studio"}, // requires callback
		"run-winetricks":   {ctl.pfx.Winetricks, "Running Winetricks"},
		"delete-prefix":    {ctl.ui.DeletePrefixes, "Clearing Data"},
		"kill-prefix":      {ctl.pfx.Kill, "Stopping Studio"},
		"uninstall-studio": {ctl.ui.Uninstall, "Uninstalling Studio"},
	}

	var label gtk.Label
	ctl.builder.GetObject("stack").Cast(&ctl.stack)

	ctl.builder.GetObject("loading-label").Cast(&label)

	for name, action := range actions {
		act := gio.NewSimpleAction(name, nil)
		actcb := func(_ gio.SimpleAction, _ uintptr) {
			ctl.stack.SetVisibleChildName("loading")
			label.SetLabel(action.msg + "...")

			if name == "run-studio" {
				b := ctl.Studio()
				action.act = b.Run
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
	ctl.builder.Unref()

	return ctl
}
