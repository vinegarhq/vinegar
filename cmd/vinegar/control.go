package main

import (
	"reflect"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/vinegarhq/vinegar/internal/adwaux"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gtkutil"
)

type control struct {
	*app

	builder *gtk.Builder
	win     adw.ApplicationWindow

	runner adw.EntryRow
}

func (s *app) newControl() *control {
	ctl := control{
		app:     s,
		builder: gtk.NewBuilderFromResource(gtkutil.Resource("ui/control.ui")),
	}

	ctl.builder.GetObject("window").Cast(&ctl.win)
	ctl.win.SetApplication(&s.Application.Application)

	var page adw.PreferencesPage
	ctl.builder.GetObject("prefpage-main").Cast(&page)
	adwaux.AddStructPage(&page, reflect.ValueOf(ctl.cfg).Elem())

	ctl.builder.GetObject("entry-prefix-run").Cast(&ctl.runner)
	applyCb := func(_ adw.EntryRow) {
		cmd := ctl.runner.GetText()
		args := strings.Fields(cmd)
		ctl.app.errThread(ctl.pfx.Wine(args[0], args[1:]...).Run)
	}
	ctl.runner.ConnectApply(&applyCb)

	var r gtk.Button
	ctl.builder.GetObject("btn-prefix-config").Cast(&r)
	cb := func(_ gtk.Button) {
		ctl.runner.SetText("winecfg")
		gobject.SignalEmitByName(&ctl.runner.Object, "apply")
	}
	r.ConnectClicked(&cb)

	for name, fn := range map[string]any{
		"save":  ctl.saveConfig,
		"about": ctl.showAbout,

		"clear-cache": ctl.clearCache,
		"open-prefix": func() {
			gtk.ShowUri(&ctl.win.Window, "file://"+ctl.pfx.Dir(), 0)
		},
		"open-logs": func() {
			gtk.ShowUri(&ctl.win.Window, "file://"+dirs.Logs, 0)
		},

		"delete-studio": ctl.deleteDeployments,
		"run":           ctl.run,
		"prefix-kill":   ctl.pfx.Kill,
		"delete-prefix": ctl.deletePrefixes,
	} {
		action := gio.NewSimpleAction(name, nil)
		activate := func(_ gio.SimpleAction, p uintptr) {
			switch v := fn.(type) {
			case func() error:
				ctl.app.errThread(func() error {
					return v()
				})
			case func():
				v()
			default:
				panic("unreachable")
			}
		}
		action.ConnectActivate(&activate)
		ctl.win.AddAction(action)
		action.Unref()
	}

	ctl.updateRun()

	return &ctl
}

func (ctl *control) updateRun() {
	var btn adw.ButtonContent
	ctl.builder.GetObject("btn-run").Cast(&btn)
	if ctl.pfx.Exists() {
		btn.SetLabel("Run Studio")
	} else {
		btn.SetLabel("Initialize")
	}
}
