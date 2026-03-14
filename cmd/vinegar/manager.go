package main

import (
	"context"
	"fmt"
	"log/slog"

	"codeberg.org/puregotk/puregotk/v4/adw"
	"codeberg.org/puregotk/puregotk/v4/gio"
	"codeberg.org/puregotk/puregotk/v4/gobject"
	"codeberg.org/puregotk/puregotk/v4/gtk"
	"github.com/google/go-github/v80/github"
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

	var tags gtk.StringList
	m.builder.GetObject("wine_tags").Cast(&tags)
	var tag gtk.DropDown
	m.builder.GetObject("wine_tag").Cast(&tag)

	var up gtk.Popover
	m.builder.GetObject("wine_updater").Cast(&up)
	upcb := func(_ gtk.Widget) {
		if tags.GetNItems() > 1 {
			return
		}
		slog.Info("Fetching releases")
		m.app.errThread(func() error {
			client := github.NewClient(nil)
			ctx := context.Background()

			releases, _, err := client.Repositories.ListReleases(
				ctx, "vinegarhq", "kombucha", nil)
			if err != nil {
				return err
			}

			for _, release := range releases {
				gutil.IdleAdd(func() {
					tags.Append(*release.TagName)
				})
			}

			return nil
		})
	}
	up.ConnectShow(&upcb)

	var row adw.ActionRow
	m.builder.GetObject("wine_row").Cast(&row)

	var dl gtk.MenuButton
	m.builder.GetObject("wine_update").Cast(&dl)

	var cnf gtk.Button
	m.builder.GetObject("wine_confirm").Cast(&cnf)
	cnfcb := func(_ gtk.Button) {
		row.SetSensitive(false)
		sel := gtk.StringObjectNewFromInternalPtr(
			tag.GetSelectedItem().Ptr)
		m.app.errThread(func() error {
			defer gutil.IdleAdd(func() {
				row.SetSensitive(true)
			})
			return a.updateWine(sel.GetString())
		})
	}
	cnf.ConnectClicked(&cnfcb)

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
