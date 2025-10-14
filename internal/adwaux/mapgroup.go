package adwaux

import (
	"maps"
	"math"
	"reflect"
	"slices"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

var kindEmpty = map[string]reflect.Value{
	// FFlag 'Log' (byte) unsupported until ???
	reflect.Bool.String():   reflect.ValueOf(true),
	reflect.String.String(): reflect.ValueOf(""),
	reflect.Int64.String():  reflect.ValueOf(int64(0)),
}

type mapGroup struct {
	mv reflect.Value

	*adw.PreferencesGroup
}

func newMapGroup(mv reflect.Value) *adw.PreferencesGroup {
	g := mapGroup{
		mv:               mv,
		PreferencesGroup: adw.NewPreferencesGroup(),
	}

	add := gtk.NewMenuButton()
	g.SetHeaderSuffix(&add.Widget)
	add.SetHalign(gtk.AlignEndValue)
	add.SetValign(gtk.AlignCenterValue)
	add.SetIconName("list-add-symbolic")
	add.AddCssClass("flat")

	new := gtk.NewPopover()
	add.SetPopover(&new.Widget)

	addName := gtk.NewEntry()
	addName.SetPlaceholderText("Key name to add")

	var addKind *gtk.DropDown
	if g.mv.Type().Elem().Kind() == reflect.Interface {
		box := gtk.NewBox(gtk.OrientationHorizontalValue, 8)
		new.SetChild(&box.Widget)
		box.Append(&addName.Widget)

		addKind = gtk.NewDropDownFromStrings(
			slices.AppendSeq(make([]string, 0, len(kindEmpty)), maps.Keys(kindEmpty)),
		)
		box.Append(&addKind.Widget)
	} else {
		new.SetChild(&addName.Widget)
	}

	activate := func(_ gtk.Entry) {
		key := reflect.ValueOf(addName.GetText())
		if g.mv.MapIndex(key).IsValid() {
			addName.AddCssClass("error")
			return
		}

		var empty reflect.Value
		if addKind != nil {
			empty = kindEmpty[gtk.StringObjectNewFromInternalPtr(
				addKind.GetSelectedItem().Ptr,
			).GetString()]
		} else {
			empty = kindEmpty["string"]
		}
		g.mv.SetMapIndex(key, empty)

		addName.SetText("")
		addName.RemoveCssClass("error")
		g.addKeyRow(key)
		g.ActivateActionVariant("win.save", nil)
		new.Hide()
	}
	addName.ConnectActivate(&activate)

	for iter := mv.MapRange(); iter.Next(); {
		g.addKeyRow(iter.Key())
	}

	return g.PreferencesGroup
}

func (g *mapGroup) addKeyRow(k reflect.Value) {
	var row *adw.PreferencesRow
	v := g.mv.MapIndex(k)
	strict := g.mv.Type().Elem().Kind() == reflect.Interface
	if strict {
		v = v.Elem()
	}

	// BUG: There is no way to delete a default variable or FFlag, since
	//      the configurations are overlayed ontop of one another.
	//      Reflection will panic if the user attempted to remove this key.
	//      I... I'm just gonna keep it here because there is no solution.

	remove := gtk.NewButton()
	remove.SetValign(gtk.AlignCenterValue)
	remove.SetIconName("edit-delete-symbolic")
	remove.AddCssClass("flat")
	clicked := func(_ gtk.Button) {
		g.mv.SetMapIndex(k, reflect.Value{})
		row.ActivateActionVariant("win.save", nil)
		g.Remove(&row.Widget)
		row.Unref()
	}
	remove.ConnectClicked(&clicked)

	switch v.Type().Kind() {
	case reflect.Bool:
		sw := adw.NewSwitchRow()
		row = &sw.PreferencesRow
		sw.AddSuffix(&remove.Widget)
		sw.SetActive(v.Bool())
		changed := func() {
			g.mv.SetMapIndex(k, reflect.ValueOf(sw.GetActive()))
			g.ActivateActionVariant("win.save", nil)
		}
		sw.ConnectSignal("notify::active", &changed)
	case reflect.String:
		entry := adw.NewEntryRow()
		row = &entry.PreferencesRow
		entry.AddSuffix(&remove.Widget)
		entry.SetText(v.String())
		if strict {
			entry.AddCssClass("monospace")
		}
		changed := func() {
			g.mv.SetMapIndex(k, reflect.ValueOf(entry.GetText()))
			g.ActivateActionVariant("win.save", nil)
		}
		entry.ConnectSignal("notify::text", &changed)
	case reflect.Int64:
		adj := gtk.NewAdjustment(0.0,
			float64(math.MinInt), float64(math.MaxInt),
			1.0, 4.0, 0.0,
		) // ;w;
		spin := adw.NewSpinRow(adj, 1, 0)
		row = &spin.PreferencesRow
		spin.AddSuffix(&remove.Widget)
		spin.SetValue(float64(v.Int()))
		changed := func() {
			g.mv.SetMapIndex(k, reflect.ValueOf(int64(spin.GetValue())))
			g.ActivateActionVariant("win.save", nil)
		}
		spin.ConnectSignal("notify::value", &changed)
	default:
		panic("adwaux: unhandled type: " + v.Type().Kind().String())
	}

	row.SetTitle(k.String())
	g.Add(&row.Widget)
}
