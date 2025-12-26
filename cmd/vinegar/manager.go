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

	handlers struct {
		// Virtual desktop rows and their spin row handler IDs
		// Signal blockers necessary as to not modify the underlying registry, only
		// modifying the view to the user.
		vdrow    uint32
		vdwidth  uint32
		vdheight uint32
	}
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

	var (
		row       adw.ExpanderRow
		widthAdj  gtk.Adjustment
		heightAdj gtk.Adjustment
	)
	m.builder.GetObject("switch-prefix-vdesktop").Cast(&row)
	m.builder.GetObject("adj-vdesktop-width").Cast(&widthAdj)
	m.builder.GetObject("adj-vdesktop-height").Cast(&heightAdj)

	expansionCb := m.virtualDesktopEnabled
	changedCb := m.adjustVirtualDesktopChanged
	m.handlers.vdrow = row.ConnectSignal("notify::enable-expansion", &expansionCb)
	m.handlers.vdwidth = widthAdj.ConnectSignal("value-changed", &changedCb)
	m.handlers.vdheight = heightAdj.ConnectSignal("value-changed", &changedCb)

	m.builder.GetObject("prefgroup-wine").Cast(&m.wine)
	m.wine.SetSensitive(m.pfx.Running())
	sensitiveCb := func() {
		if m.wine.GetSensitive() {
			m.setVirtualDesktopAdjustment()
		}
		slog.Info("Running", m.wine.GetSensitive(), m.pfx.Running())
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

func (m *manager) adjustVirtualDesktopChanged() {
	var (
		widthAdj  gtk.Adjustment
		heightAdj gtk.Adjustment
	)
	m.builder.GetObject("adj-vdesktop-width").Cast(&widthAdj)
	m.builder.GetObject("adj-vdesktop-height").Cast(&heightAdj)

	res := fmt.Sprintf("%.0fx%.0f", widthAdj.GetValue(), heightAdj.GetValue())

	slog.Info("Changing Virtual Desktop resolution", "res", res)
	err := m.pfx.RegistryAdd(`HKCU\Software\Wine\Explorer\Desktops`, "Default", res)
	if err != nil {
		m.showError(fmt.Errorf("virtual desktop set: %w", err))
	}
}

func (m *manager) virtualDesktopEnabled() {
	var row adw.ExpanderRow
	m.builder.GetObject("switch-prefix-vdesktop").Cast(&row)
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
		m.adjustVirtualDesktopChanged()
	}
}

func (m *manager) setVirtualDesktopAdjustment() {
	var (
		row       adw.ExpanderRow
		widthAdj  gtk.Adjustment
		heightAdj gtk.Adjustment
	)
	m.builder.GetObject("switch-prefix-vdesktop").Cast(&row)
	m.builder.GetObject("adj-vdesktop-width").Cast(&widthAdj)
	m.builder.GetObject("adj-vdesktop-height").Cast(&heightAdj)

	k, err := m.pfx.RegistryQuery(`HKCU\Software\Wine\Explorer\Desktops`)
	if err != nil {
		m.showError(fmt.Errorf("virtual desktop state: %w", err))
		return
	}
	gobject.SignalHandlerBlock(&row.Object, m.handlers.vdrow)
	defer func() { gobject.SignalHandlerUnblock(&row.Object, m.handlers.vdrow) }()
	if k == nil {
		row.SetEnableExpansion(false)
		return
	}

	gobject.SignalHandlerBlock(&widthAdj.Object, m.handlers.vdwidth)
	gobject.SignalHandlerBlock(&heightAdj.Object, m.handlers.vdheight)

	var width, height float64
	fmt.Sscanf(k.GetValue("Default").Data.(string), "%fx%f", &width, &height)
	if widthAdj.GetValue() != width {
		widthAdj.SetValue(width)
	}
	if heightAdj.GetValue() != height {
		heightAdj.SetValue(height)
	}
	row.SetEnableExpansion(true)

	gobject.SignalHandlerUnblock(&widthAdj.Object, m.handlers.vdwidth)
	gobject.SignalHandlerUnblock(&heightAdj.Object, m.handlers.vdheight)
}
