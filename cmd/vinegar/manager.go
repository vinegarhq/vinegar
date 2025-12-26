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
	sensitiveCb := func() {
		if m.wine.GetSensitive() {
			m.setupVirtualDesktopAdjustment()
		}
	}
	m.wine.ConnectSignal("notify::sensitive", &sensitiveCb)
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

	gutil.IdleAdd(func() {
		toast := adw.NewToast(s)
		toast.SetPriority(adw.ToastPriorityHighValue)
		overlay.AddToast(toast)
	})
}

func (m *manager) startWine() error {
	defer func() {
		if m.pfx.Running() {
			m.wine.SetSensitive(true)
			m.showToast("Wine initialized")
		}
	}()
	return m.pfx.Start()
}

func (m *manager) setupVirtualDesktopAdjustment() {
	var (
		row       adw.ExpanderRow
		widthAdj  gtk.Adjustment
		heightAdj gtk.Adjustment
	)
	m.builder.GetObject("switch-prefix-vdesktop").Cast(&row)
	m.builder.GetObject("adj-vdesktop-width").Cast(&widthAdj)
	m.builder.GetObject("adj-vdesktop-height").Cast(&heightAdj)
	expansionCb := func() {
		enable := row.GetEnableExpansion()

		slog.Info("Enabling Virtual Desktop", "state", enable)
		var err error
		if enable {
			err = m.pfx.RegistryAdd(`HKCU\Software\Wine\Explorer`, "Desktop", "Default")
		} else {
			err = m.pfx.RegistryDelete(`HKCU\Software\Wine\Explorer`, "")
		}
		if err != nil {
			m.showError(fmt.Errorf("virtual desktop adj: %w", err))
		} else if enable { // If enabling it fails, then setting would probably fail
			gobject.SignalEmitByName(&widthAdj.Object, "value-changed")
		}
	}
	changedCb := func() {
		res := fmt.Sprintf("%.0fx%.0f", widthAdj.GetValue(), heightAdj.GetValue())

		slog.Info("Changing Virtual Desktop resolution", "res", res)
		err := m.pfx.RegistryAdd(`HKCU\Software\Wine\Explorer\Desktops`, "Default", res)
		if err != nil {
			m.showError(fmt.Errorf("virtual desktop set: %w", err))
		}
	}

	row.SetSensitive(false)

	m.errThread(func() error {
		defer func() {
			row.SetSensitive(true)
		}()
		k, err := m.pfx.RegistryQuery(`HKCU\Software\Wine\Explorer\Desktops`)
		if err != nil {
			return fmt.Errorf("virtual desktop state: %w", err)
		}
		if k == nil {
			row.SetEnableExpansion(false)
			return nil
		}

		var width, height int
		fmt.Sscanf(k.GetValue("Default").Data.(string), "%dx%d", &width, &height)
		widthAdj.SetValue(float64(width))
		heightAdj.SetValue(float64(height))
		row.SetEnableExpansion(true)
		return nil
	})

	widthAdj.ConnectSignal("value-changed", &changedCb)
	heightAdj.ConnectSignal("value-changed", &changedCb)
	row.ConnectSignal("notify::enable-expansion", &expansionCb)
}
