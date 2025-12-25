package main

import (
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/vinegarhq/vinegar/internal/adwaux"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gutil"
)

type manager struct {
	*app

	builder *gtk.Builder
	win     adw.ApplicationWindow

	runner adw.EntryRow
}

func (a *app) newManager() *manager {
	m := manager{
		app:     a,
		builder: gtk.NewBuilderFromResource(gutil.Resource("ui/manager.ui")),
	}

	m.builder.GetObject("window").Cast(&m.win)
	m.win.SetApplication(&a.Application.Application)

	var page adw.PreferencesPage
	m.builder.GetObject("prefpage-main").Cast(&page)
	adwaux.AddStructPage(&page, reflect.ValueOf(m.cfg).Elem())

	m.builder.GetObject("entry-prefix-run").Cast(&m.runner)
	applyCb := func(_ adw.EntryRow) {
		cmd := m.runner.GetText()
		args := strings.Fields(cmd)
		if len(args) < 1 {
			return
		}
		slog.Info("Running Wine command", "args", args)
		m.app.errThread(m.pfx.Wine(args[0], args[1:]...).Run)
	}
	m.runner.ConnectApply(&applyCb)

	var (
		vdesktop adw.ExpanderRow
		width    gtk.Adjustment
		height   gtk.Adjustment
	)
	m.builder.GetObject("switch-prefix-vdesktop").Cast(&vdesktop)
	m.builder.GetObject("adj-vdesktop-width").Cast(&width)
	m.builder.GetObject("adj-vdesktop-height").Cast(&height)
	expansionCb := func() {
		enable := vdesktop.GetEnableExpansion()
		slog.Info("Enabling Virtual Desktop", "state", enable)
		if vdesktop.GetEnableExpansion() {
			gobject.SignalEmitByName(&width.Object, "value-changed")
		}

		var err error
		if enable {
			err = m.pfx.RegistryAdd(`HKCU\Software\Wine\Explorer`, "Desktop", "Default")
		} else {
			err = m.pfx.RegistryDelete(`HKCU\Software\Wine\Explorer`, "")
		}
		if err != nil {
			m.showError(err)
		}
	}
	changedCb := func() {
		res := fmt.Sprintf("%.0fx%.0f", width.GetValue(), height.GetValue())

		slog.Info("Changing Virtual Desktop resolution", "res", res)

		err := m.pfx.RegistryAdd(`HKCU\Software\Wine\Explorer\Desktops`, "Default", res)
		if err != nil {
			m.showError(err)
		}
	}
	width.ConnectSignal("value-changed", &changedCb)
	height.ConnectSignal("value-changed", &changedCb)
	vdesktop.ConnectSignal("notify::enable-expansion", &expansionCb)

	var r gtk.Button
	m.builder.GetObject("btn-prefix-config").Cast(&r)
	cb := func(_ gtk.Button) {
		m.runner.SetText("winecfg")
		gobject.SignalEmitByName(&m.runner.Object, "apply")
	}
	r.ConnectClicked(&cb)

	for name, fn := range map[string]any{
		"save":  m.saveConfig,
		"about": m.showAbout,
		"open-prefix": func() {
			gtk.ShowUri(&m.win.Window, "file://"+m.pfx.Dir(), 0)
		},
		"open-logs": func() {
			gtk.ShowUri(&m.win.Window, "file://"+dirs.Logs, 0)
		},
		"run": m.run,

		"prefix-kill":   m.killPrefix,
		"delete-prefix": m.deletePrefixes,
		"delete-studio": m.deleteDeployments,
		"clear-cache":   m.clearCache,
	} {
		action := gio.NewSimpleAction(name, nil)
		activate := func(_ gio.SimpleAction, p uintptr) {
			switch v := fn.(type) {
			case func() error:
				m.app.errThread(func() error {
					return v()
				})
			case func():
				v()
			default:
				panic("unreachable")
			}
		}
		action.ConnectActivate(&activate)
		m.win.AddAction(action)
		action.Unref()
	}

	m.updateRun()

	return &m
}

func (m *manager) updateRun() {
	var btn adw.ButtonContent
	m.builder.GetObject("btnc-run").Cast(&btn)
	btn.SetIconName("media-playback-start-symbolic")
	if len(m.boot.procs) > 0 {
		btn.SetIconName("media-playback-stop-symbolic")
		btn.SetLabel("Stop")
		return
	}
	if m.pfx.Exists() {
		btn.SetLabel("Run Studio")
	} else {
		btn.SetLabel("Initialize")
	}
}

func (m *manager) showToast(s string) {
	var overlay adw.ToastOverlay
	m.builder.GetObject("overlay").Cast(&overlay)

	gutil.IdleAdd(func() {
		toast := adw.NewToast(s)
		overlay.AddToast(toast)
	})
}
