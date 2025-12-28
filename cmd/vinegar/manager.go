package main

import (
	"fmt"
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
}

func (a *app) newManager() *manager {
	m := manager{
		app: a,
	}

	m.builder = gtk.NewBuilderFromResource(gutil.Resource("ui/manager.ui"))
	m.builder.GetObject("window").Cast(&m.win)
	m.win.SetApplication(&a.Application.Application)

	var cmd gtk.Entry
	m.builder.GetObject("entry-cmd").Cast(&cmd)
	cb := m.runWineCmd
	cmd.ConnectActivate(&cb)

	var page adw.PreferencesPage
	m.builder.GetObject("prefpage-main").Cast(&page)
	adwaux.AddStructPage(&page, reflect.ValueOf(m.cfg).Elem())

	for name, fn := range map[string]any{
		"save":  m.saveConfig,
		"about": m.showAbout,
		"open-prefix": func() {
			gtk.ShowUri(&m.win.Window, "file://"+m.pfx.Dir(), 0)
		},
		"open-logs": func() {
			gtk.ShowUri(&m.win.Window, "file://"+dirs.Logs, 0)
		},

		"prefix-kill":   m.killPrefix,
		"delete-prefix": m.deletePrefixes,
		"delete-studio": m.deleteDeployments,
		"clear-cache":   m.clearCache,
		"update":        m.updateWine,

		"winecfg": func() {
			cmd.SetText("winecfg")
			gobject.SignalEmitByName(&cmd.Object, "activate")
		},
	} {
		m.addAction(name, fn)
	}

	run := gio.NewSimpleAction("run", nil)
	activate := func(_ gio.SimpleAction, p uintptr) {
		m.app.errThread(m.run)
	}
	run.ConnectActivate(&activate)
	m.win.AddAction(run)
	run.Unref()
	m.updateRunContent()

	return &m
}

func (m *manager) addAction(name string, fn any) {
	action := gio.NewSimpleAction(name, nil)
	activate := func(_ gio.SimpleAction, p uintptr) {
		switch v := fn.(type) {
		case func() error:
			m.app.errThread(func() error {
				defer m.loading()()
				return v()
			})
		case func():
			v()
		default:
			panic(fmt.Sprintf("unhandled function type: %T", fn))
		}
	}
	action.ConnectActivate(&activate)
	m.win.AddAction(action)
	action.Unref()
}

func (m *manager) loading() func() {
	var button gtk.Button
	var bar adw.HeaderBar

	m.builder.GetObject("session").Cast(&button)
	m.builder.GetObject("headerbar").Cast(&bar)

	spinner := adw.NewSpinner()

	gutil.IdleAdd(func() {
		button.SetSensitive(false)
		bar.PackStart(&spinner.Widget)
		bar.SetMarginStart(6)
		bar.SetMarginEnd(6)
	})

	return func() {
		gutil.IdleAdd(func() {
			bar.Remove(&spinner.Widget)
			button.SetSensitive(true)
			m.updateRunContent()
		})
	}
}

func (m *manager) updateRunContent() {
	var wine adw.HeaderBar
	m.builder.GetObject("prefgroup-wine").Cast(&wine)
	wine.SetSensitive(m.pfx.Exists())

	var c adw.ButtonContent
	m.builder.GetObject("session-content").Cast(&c)
	c.SetIconName("media-playback-start-symbolic")
	if len(m.boot.procs) > 0 {
		c.SetIconName("media-playback-stop-symbolic")
		c.SetLabel("Stop")
		return
	}
	if m.pfx.Exists() {
		c.SetLabel("Run Studio")
	} else {
		c.SetLabel("Initialize")
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
