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
	m.wine.SetSensitive(m.pfx.Exists())

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
		"update":        m.updateWine,
	} {
		action := gio.NewSimpleAction(name, nil)
		activate := func(_ gio.SimpleAction, p uintptr) {
			stop := m.loading()
			switch v := fn.(type) {
			case func() error:
				m.app.errThread(func() error {
					defer stop()
					return v()
				})
			case func():
				v()
				stop()
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

func (m *manager) loading() func() {
	var bar adw.HeaderBar
	m.builder.GetObject("headerbar").Cast(&bar)
	spinner := adw.NewSpinner()
	bar.PackStart(&spinner.Widget)
	bar.SetMarginStart(6)
	bar.SetMarginEnd(6)
	return func() {
		bar.Remove(&spinner.Widget)
	}
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
	if m == nil {
		return
	}
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
	slog.Info("Starting Wineserver")
	return m.pfx.Start()
}
