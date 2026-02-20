package main

import (
	"fmt"
	"log/slog"
	"math"
	"slices"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/dirs"
	"github.com/vinegarhq/vinegar/internal/gutil"
	"github.com/vinegarhq/vinegar/internal/sysinfo"
)

func (m *manager) connectElements() {
	b := m.builder
	cfg := &m.cfg.Studio

	// Currently, every possible configuration value, on the proper signal
	// that modifies the value that represents the value for a property,
	// is to be swap connected to win.saveCb on that specific signal, as
	// seen in [connectSave].
	var saveCb gobject.Callback = func() {
		m.win.ActivateActionVariant("win.save", nil)
	}
	connectSave := func(obj *gobject.Object, signal string) {
		gobject.SignalConnectData(obj, signal, &saveCb,
			0, nil, gobject.GConnectSwappedValue)
	}

	signalSave := func(w *gtk.Widget, signal string, callback func()) {
		gutil.ConnectSignal(w, signal, callback)
		connectSave(&w.Object, signal)
	}
	simpleEntry := func(name string, setting *string) {
		entry := gutil.GetObject[adw.EntryRow](b, name)
		entry.SetText(*setting)
		signalSave(&entry.Widget, "apply", func() {
			*setting = entry.GetText()
		})
	}
	simpleSwitch := func(name string, setting *bool) {
		row := gutil.GetObject[adw.SwitchRow](b, name)
		row.SetActive(*setting)
		signalSave(&row.Widget, "notify::active", func() {
			*setting = row.GetActive()
		})
	}

	// Function bindings to their configuration element in the UI is
	// in the order of their appearance in the UI.

	web := gutil.GetObject[adw.SwitchRow](b, "webpages_row")
	web.SetActive(cfg.WebView != "")
	signalSave(&web.Widget, "notify::active", func() {
		if web.GetActive() {
			cfg.WebView = config.WebViewVersion
		} else {
			cfg.WebView = ""
		}
	})

	wine := gutil.GetObject[adw.ActionRow](b, "wine_row")
	wine.SetSubtitle(cfg.WineRoot)
	signalSave(&wine.Widget, "notify::subtitle", func() {
		slog.Info("Wineset")
		cfg.WineRoot = wine.GetSubtitle()
	})
	gutil.ConnectBuilderSimple(b, "wine_revert", "clicked", func() {
		if cfg.WineRoot != "" {
			cfg.WineRoot = ""
		} else {
			cfg.WineRoot = dirs.WinePath
		}
		wine.SetSubtitle(cfg.WineRoot)
	})
	gutil.ConnectBuilderSimple(b, "wine_select", "clicked", func() {
		dialog := gtk.NewFileDialog()
		var ready gio.AsyncReadyCallback = func(_, resPtr, _ uintptr) {
			res := gio.SimpleAsyncResultNewFromInternalPtr(resPtr)
			f, err := dialog.SelectFolderFinish(res)
			if err != nil {
				slog.Error("FileDialog error", "err", err)
				return
			}
			wine.SetSubtitle(f.GetPath())
		}
		win := gtk.WindowNewFromInternalPtr(wine.GetRoot().Ptr)
		dialog.SelectFolder(win, nil, &ready, 0)
	})

	// UI model MUST represent the same values by index.
	renderer := gutil.GetObject[adw.ComboRow](b, "renderer_row")
	renderer.SetSelected(uint(slices.Index(config.RendererValues, cfg.Renderer)))
	signalSave(&renderer.Widget, "notify::selected-item", func() {
		cfg.Renderer = config.RendererValues[renderer.GetSelected()]
	})

	desktop := gutil.GetObject[adw.ExpanderRow](b, "desktop_row")
	desktop.SetEnableExpansion(cfg.Desktop != "")
	resolution := gutil.GetObject[adw.EntryRow](b, "resolution_entry")
	resolution.SetText(cfg.Desktop)
	signalSave(&resolution.Widget, "notify::text", func() {
		cfg.Desktop = resolution.GetText()
		if cfg.Desktop == "" {
			desktop.SetExpanded(false)
			desktop.SetEnableExpansion(false)
		}
	})
	gutil.ConnectSignal(&desktop.Widget, "notify::enable-expansion", func() {
		// Automatically sets the feature to disabled and
		// unexpands as seen in notify::text, when properly
		// disabled whether switch or cleared.
		if !desktop.GetEnableExpansion() && !desktop.GetExpanded() {
			resolution.SetText("")
		} else if resolution.GetText() == "" {
			resolution.SetText(config.DesktopsResolution)
		}
	})

	card := gutil.GetObject[adw.ComboRow](b, "cards_row")
	cards := gutil.GetObject[gtk.StringList](b, "cards")
	values := make(map[string]string, len(sysinfo.Cards))
	for i, c := range sysinfo.Cards {
		if c.String() == cfg.ForcedGpu {
			card.SetSelected(uint(i))
		}
		shown := fmt.Sprintf("%d: %s", c.Index, c.Product)
		values[shown] = c.String()
		cards.Append(shown)
	}
	signalSave(&card.Widget, "notify::selected-item", func() {
		cfg.ForcedGpu = values[cards.GetString(card.GetSelected())]
	})

	simpleEntry("launcher_row", &cfg.Launcher)

	simpleSwitch("discord_row", &cfg.DiscordRPC)
	simpleSwitch("gamemode_row", &cfg.GameMode)

	env := gutil.GetObject[adw.ExpanderRow](b, "env_row")
	for key := range cfg.Env {
		addKeyRow(&env, cfg.Env, key)
	}
	envPopover := gutil.GetObject[gtk.Popover](b, "env_popover")
	gutil.ConnectBuilder[gtk.Entry](b, "env_entry", "activate", func(entry *gtk.Entry) {
		key := entry.GetText()

		if _, ok := cfg.Env[key]; ok {
			entry.AddCssClass("error")
			return
		}
		entry.RemoveCssClass("error")

		// Ensure a value is present for addKeyRow, user is expected to
		// provide some sort of value.
		cfg.Env[key] = ""
		addKeyRow(&env, cfg.Env, key)

		// [1]: Incase user just wanted to add a variable without touching the value
		entry.ActivateActionVariant("win.save", nil)
		envPopover.Hide()
	})

	fflags := gutil.GetObject[adw.ExpanderRow](b, "fflags_row")
	for key := range cfg.FFlags {
		addKeyRow(&fflags, cfg.FFlags, key)
	}
	newFFlagType := gutil.GetObject[gtk.DropDown](b, "fflag_type")
	fflagPopover := gutil.GetObject[gtk.Popover](b, "fflag_popover")
	gutil.ConnectBuilder[gtk.Entry](b, "fflag_name", "activate", func(entry *gtk.Entry) {
		key := entry.GetText()

		if _, ok := cfg.FFlags[key]; ok {
			entry.AddCssClass("error")
			return
		}
		entry.RemoveCssClass("error")

		selected := gtk.StringObjectNewFromInternalPtr(newFFlagType.GetSelectedItem().Ptr)
		// Ensure to create default values for these types when adding a new
		// FFlag, required to represent any possible value, which the user
		// then is expected to edit to a value of their choosing.
		switch selected.GetString() {
		case "Number":
			cfg.FFlags[key] = int64(0)
		case "Boolean":
			cfg.FFlags[key] = false
		case "String":
			cfg.FFlags[key] = ""
		}
		addKeyRow(&fflags, cfg.FFlags, key)

		entry.ActivateActionVariant("win.save", nil) // [1]
		fflagPopover.Hide()
	})

	simpleEntry("version_row", &cfg.ForcedVersion)
	simpleEntry("channel_row", &cfg.Channel)
}

// addKeyRow makes a new custom widget that represents the key value
// pair [key] for the given map at [m], by asserting the existing type
// for the key in the map, and making an appropiate switch, spin, or entry
// row for the key, and adding it to the given expander row at [w].
func addKeyRow[V any](
	w *adw.ExpanderRow,
	m map[string]V,
	key string,
) {
	var row *adw.PreferencesRow

	remove := gtk.NewButton()
	remove.SetValign(gtk.AlignCenterValue)
	remove.SetIconName("edit-delete-symbolic")
	remove.AddCssClass("flat")
	gutil.ConnectSignal(remove, "clicked", func() {
		delete(m, row.GetTitle())
		row.ActivateActionVariant("win.save", nil)
		w.Remove(&row.Widget)
		row.Unref()
	})

	val, ok := m[key]
	if !ok {
		panic(fmt.Sprintf("bind: key %s must already exist", key))
	}
	// any($toset).(V) required to set a value on a map in generic
	switch val := any(val).(type) {
	case bool:
		sw := adw.NewSwitchRow()
		sw.AddSuffix(&remove.Widget)
		sw.SetActive(val)
		gutil.ConnectSignal(sw, "notify::active", func() {
			m[key] = any(sw.GetActive()).(V)
			sw.ActivateActionVariant("win.save", nil)
		})
		row = &sw.PreferencesRow
	case string:
		entry := adw.NewEntryRow()
		entry.AddSuffix(&remove.Widget)
		entry.SetText(val)
		entry.AddCssClass("monospace")
		gutil.ConnectSignal(entry, "notify::text", func() {
			m[key] = any(entry.GetText()).(V)
			entry.ActivateActionVariant("win.save", nil)
		})
		row = &entry.PreferencesRow
	case int64:
		adj := gtk.NewAdjustment(0.0,
			float64(math.MinInt), float64(math.MaxInt),
			1.0, 4.0, 0.0,
		) // ;w;
		spin := adw.NewSpinRow(adj, 1, 0)
		spin.AddSuffix(&remove.Widget)
		spin.SetValue(float64(val))
		gutil.ConnectSignal(spin, "notify::value", func() {
			m[key] = any(int64(spin.GetValue())).(V)
			spin.ActivateActionVariant("win.save", nil)
		})
		row = &spin.PreferencesRow
	default:
		panic(fmt.Sprintf("bind: unhandled type %T", val))
	}

	row.SetTitle(key)
	w.SetExpanded(true)
	w.AddRow(&row.Widget)
	return
}
