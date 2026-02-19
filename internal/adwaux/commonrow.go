package adwaux

import (
	"maps"
	"reflect"
	"slices"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

func newComboRow(v reflect.Value, values map[string]string) *adw.ComboRow {
	// TODO: Should use a GtkExpression for this.
	combo := adw.NewComboRow()
	model := gtk.NewStringList(nil)

	for i, name := range slices.Collect(maps.Keys(values)) {
		if v.String() == values[name] {
			defer combo.SetSelected(uint(i))
		}
		model.Append(name)
	}
	// Prefer implicit "None" in Selector implementation
	combo.SetSelected(0)

	combo.SetModel(model)
	selectedItem := func() {
		v.SetString(values[model.GetString(combo.GetSelected())])
		combo.ActivateActionVariant("win.save", nil)
	}
	combo.ConnectSignal("notify::selected-item", &selectedItem)
	return combo
}

func newEntryRow(v reflect.Value) *adw.EntryRow {
	ent := adw.NewEntryRow()
	ent.SetText(v.String())
	ent.SetShowApplyButton(true)
	apply := func(_ adw.EntryRow) {
		v.SetString(ent.GetText())
		ent.ActivateActionVariant("win.save", nil)
	}
	ent.ConnectApply(&apply)
	return ent
}

func newSwitchRow(v reflect.Value) *adw.SwitchRow {
	sw := adw.NewSwitchRow()
	sw.SetActive(v.Bool())
	activate := func() {
		v.SetBool(!v.Bool())
		sw.ActivateActionVariant("win.save", nil)
	}
	sw.ConnectSignal("notify::active", &activate)
	return sw
}
