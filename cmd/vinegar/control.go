package main

import (
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

func (ctl *control) Studio() {
	b := ctl.NewBootstrapper()

	destroy := func(_ gtk.Window) bool {
		// BUG: realistically no other way to cancel
		//      the bootstrapper, so just exit immediately!
		b.app.Quit()
		return false
	}
	b.win.ConnectCloseRequest(&destroy)

	b.Start()
}

func (s *ui) NewControl() control {
	ctl := control{
		ui:      s,
		builder: gtk.NewBuilderFromString(resource("control.ui"), -1),
	}

	ctl.builder.GetObject("window").Cast(&ctl.win)
	ctl.win.SetApplication(&s.app.Application)

	actions := map[string]struct {
		act func()
		msg string
	}{
		"run-studio":     {ctl.Studio, "Running Studio"},
		"run-winetricks": {nil, "Running Winetricks"},
		"delete-prefix":  {nil, "Clearing Data"},
		"uninstall":      {nil, "Uninstalling Studio"},
	}

	var label gtk.Label
	ctl.builder.GetObject("stack").Cast(&ctl.stack)

	ctl.builder.GetObject("loading-label").Cast(&label)

	for name, action := range actions {
		act := gio.NewSimpleAction(name, nil)
		actcb := func(_ gio.SimpleAction, _ uintptr) {
			ctl.stack.SetVisibleChildName("loading")
			label.SetLabel(action.msg + "...")
			action.act()
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
