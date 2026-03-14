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

	view := gutil.GetObject[adw.NavigationView](m.builder, "navigation")
	if !a.pfx.Exists() {
		view.PushByTag("welcome")
	}

	cmd := gutil.GetObject[gtk.Entry](m.builder, "cmd")
	cb := m.runWineCmd
	cmd.ConnectActivate(&cb)

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

	tags := gutil.GetObject[gtk.StringList](m.builder, "wine_tags")
	gutil.ConnectBuilderSimple(m.builder, "wine_updater", "show", func() {
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
	})

	wineRow := gutil.GetObject[adw.ActionRow](m.builder, "wine_row")
	updateWine := gutil.GetObject[gtk.Button](m.builder, "wine_confirm")
	selectedTag := gutil.GetObject[gtk.DropDown](m.builder, "wine_tag")
	cnfcb := func(_ gtk.Button) {
		wineRow.SetSensitive(false)
		sel := gtk.StringObjectNewFromInternalPtr(
			selectedTag.GetSelectedItem().Ptr)
		m.app.errThread(func() error {
			defer gutil.IdleAdd(func() {
				wineRow.SetSensitive(true)
			})
			return a.updateWine(sel.GetString())
		})
	}
	updateWine.ConnectClicked(&cnfcb)

	return &m
}

func (m *manager) showToast(s string) {
	if m == nil {
		return
	}
	overlay := gutil.GetObject[adw.ToastOverlay](m.builder, "overlay")
	slog.Info(s)

	gutil.IdleAdd(func() {
		toast := adw.NewToast(s)
		toast.SetPriority(adw.ToastPriorityHighValue)
		overlay.AddToast(toast)
	})
}
