package adwaux

import (
	"reflect"

	"github.com/jwijenbergh/puregotk/v4/adw"
)

func newOptionEntryRow(v reflect.Value, valueTitle, defaultValue string) *adw.ExpanderRow {
	exp := adw.NewExpanderRow()
	exp.SetShowEnableSwitch(true)
	exp.SetEnableExpansion(v.String() != "")

	entry := adw.NewEntryRow()
	entry.SetTitle(valueTitle)
	entry.SetText(v.String())

	text := func() {
		v.SetString(entry.GetText())
		if v.String() == "" {
			exp.SetExpanded(false)
			exp.SetEnableExpansion(false)
		}
		entry.ActivateActionVariant("win.save", nil)
	}
	entry.ConnectSignal("notify::text", &text)
	exp.AddRow(&entry.Widget)

	ee := func() {
		if !exp.GetEnableExpansion() && !exp.GetExpanded() {
			entry.SetText("")
		} else if entry.GetText() == "" {
			entry.SetText(defaultValue)
		}
	}
	exp.ConnectSignal("notify::enable-expansion", &ee)

	return exp
}
