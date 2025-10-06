// I would rather do this than define a custom GObject using CGo and generate
// all the GIR and signals responsible for that parent type.
package adwaux

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/adw"
)

// Transform value names to their Title Case counterparts
var configNameExp = regexp.MustCompile(`([a-z])([A-Z])`)

type StructPage struct {
	sv     reflect.Value
	groups map[string]*adw.PreferencesGroup

	*adw.PreferencesPage
}

func AddStructPage(page *adw.PreferencesPage, sv reflect.Value) {
	p := StructPage{
		sv:              sv,
		groups:          make(map[string]*adw.PreferencesGroup),
		PreferencesPage: page,
	}

	for _, sf := range reflect.VisibleFields(sv.Type()) {
		v := sv.FieldByIndex(sf.Index)
		p.addField(sf, v)
	}
}

func (p *StructPage) addField(sf reflect.StructField, v reflect.Value) {
	if v.Kind() == reflect.Struct {
		for _, vsf := range reflect.VisibleFields(v.Type()) {
			vv := v.FieldByIndex(vsf.Index)
			p.addField(vsf, vv)
		}
		return
	}

	groupName := sf.Tag.Get("group")
	if groupName == "" {
		panic("adwaux: expected group name for " + sf.Name)
	}

	var group *adw.PreferencesGroup
	var ok bool
	if v.Kind() == reflect.Map {
		group = newMapGroup(v)
		group.SetTitle(groupName)
		p.PreferencesPage.Add(group)
	} else if group, ok = p.groups[groupName]; !ok {
		group = adw.NewPreferencesGroup()
		p.groups[groupName] = group
		group.SetTitle(groupName)
		p.PreferencesPage.Add(group)
	}

	if v.Kind() == reflect.Map {
		return // MapKeyGroup was initialized
	}

	title := configNameExp.ReplaceAllString(sf.Name, `$1 $2`)
	fields := strings.Split(sf.Tag.Get("row"), ",")
	description := fields[0]
	option := ""
	if len(fields) > 1 {
		option = fields[1]
	}

	switch k := v.Kind(); {

	case option == "vals":
		combo := newComboRow(v, fields[2:])
		combo.SetTitle(title)
		combo.SetSubtitle(description)
		group.Add(&combo.Widget)
	case option == "path":
		path := newPathRow(v)
		path.SetTitle(description)
		group.Add(&path.Widget)

	case k == reflect.Bool:
		sw := newSwitchRow(v)
		sw.SetTitle(title)
		sw.SetSubtitle(description)
		group.Add(&sw.Widget)
	case k == reflect.String:
		ent := newEntryRow(v)
		ent.SetTitle(description)
		group.Add(&ent.Widget)

	default:
		panic("adwaux: unhandled type " + v.Kind().String() + " from " + sf.Name)
	}
}
