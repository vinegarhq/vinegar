package main

import (
	"fmt"
	"log/slog"

	"codeberg.org/puregotk/puregotk/v4/adw"
	"codeberg.org/puregotk/puregotk/v4/gio"
	"codeberg.org/puregotk/puregotk/v4/gobject"
	"codeberg.org/puregotk/puregotk/v4/gtk"
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

	var view adw.NavigationView
	m.builder.GetObject("navigation").Cast(&view)
	if !a.pfx.Exists() {
		view.PushByTag("welcome")
	}

	var cmd gtk.Entry
	m.builder.GetObject("cmd").Cast(&cmd)
	cb := m.runWineCmd
	cmd.ConnectActivate(&cb)

	var page adw.PreferencesPage
	m.builder.GetObject("main-page").Cast(&page)

	m.connectElements()
	// adwaux.AddStructGroups(&page, a.cfg)

	for name, fn := range map[string]any{
		"save":  m.saveConfig,
		"about": m.showAbout,
		"open-data": func() {
			gtk.ShowUri(&m.win.Window, "file://"+dirs.Data, 0)
		},
		"open-logs": func() {
			gtk.ShowUri(&m.win.Window, "file://"+dirs.Logs, 0)
		},

		"prefix-kill":   m.killPrefix,
		"delete-prefix": m.deletePrefixes,
		"delete-studio": m.deleteDeployments,
		"clear-cache":   m.clearCache,
		"update":        m.updateWine,
		"restore":       m.boot.restoreSettings,

		"winecfg": func() {
			cmd.SetText("winecfg")
			gobject.SignalEmitByName(&cmd.Object, "activate")
		},
	} {
		m.addAction(name, fn)
	}

	return &m
}

func (m *manager) addAction(name string, fn any) {
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
			panic(fmt.Sprintf("unhandled function type: %T", fn))
		}
	}
	action.ConnectActivate(&activate)
	m.win.AddAction(action)
	action.Unref()
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
