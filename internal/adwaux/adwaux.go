// I would rather do this than define a custom GObject using CGo and generate
// all the GIR and signals responsible for that parent type.
package adwaux

import (
	"reflect"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/adw"
)

// Enum reflection is impossible without an interface to get
// the enum values
type Selector interface {
	Values() []string
}

type PathSelector interface {
	SelectPath() // Stub function, only declares that type is path
	Defaulter
}

type Defaulter interface {
	Default() string
}

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

	groupName, ok := sf.Tag.Lookup("group")
	if !ok {
		panic("adwaux: expected group name for " + sf.Name)
	}
	if groupName == "hidden" {
		return
	}

	var group *adw.PreferencesGroup
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

	title := sf.Tag.Get("title")
	if title == "" {
		title = sf.Name
	}

	fields := strings.Split(sf.Tag.Get("row"), ",")
	description := fields[0]
	option := ""
	if len(fields) > 1 {
		option = fields[1]
	}

	if p, ok := reflect.TypeAssert[PathSelector](v); ok {
		path := newPathRow(v, p.Default())
		path.SetTitle(description)
		group.Add(&path.Widget)
	} else if t, ok := reflect.TypeAssert[Defaulter](v); ok {
		opt := newOptionEntryRow(v, fields[1], t.Default())
		opt.SetTitle(title)
		opt.SetSubtitle(description)
		group.Add(&opt.Widget)
		return
	} else if s, ok := reflect.TypeAssert[Selector](v); ok {
		combo := newComboRow(v, s.Values())
		combo.SetTitle(title)
		combo.SetSubtitle(description)
		group.Add(&combo.Widget)
		return
	}

	switch k := v.Kind(); {
	case option == "path":

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
