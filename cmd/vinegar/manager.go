package main

import (
	"log/slog"
	"reflect"

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
	wine   adw.PreferencesGroup
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
		m.errThread(m.runWineCmd)
	}
	m.runner.ConnectApply(&applyCb)

	var r gtk.Button
	m.builder.GetObject("btn-prefix-config").Cast(&r)
	cb := func(_ gtk.Button) {
		m.runner.SetText("winecfg")
		gobject.SignalEmitByName(&m.runner.Object, "apply")
	}
	r.ConnectClicked(&cb)

	m.builder.GetObject("prefgroup-wine").Cast(&m.wine)
	m.wine.SetSensitive(m.pfx.Running())
	m.errThread(m.startWine)

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
	slog.Info(s)

	gutil.IdleAdd(func() {
		toast := adw.NewToast(s)
		toast.SetPriority(adw.ToastPriorityHighValue)
		overlay.AddToast(toast)
	})
}

func (m *manager) startWine() error {
	gutil.IdleAdd(func() {
		m.wine.SetSensitive(false)
	})

	defer func() {
		if !m.pfx.Running() {
			return // Error occured
		}
		gutil.IdleAdd(func() {
			m.showToast("Wine initialized")
			m.wine.SetSensitive(true)
		})
	}()
	slog.Info("Starting Wine")
	return m.pfx.Start()
}
