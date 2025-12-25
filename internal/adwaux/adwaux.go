// I would rather do this than define a custom GObject using CGo and generate
// all the GIR and signals responsible for that parent type.
package adwaux

import (
	"reflect"
	"strings"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/gtk"

	. "github.com/pojntfx/go-gettext/pkg/i18n"
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

type Groups map[string]*adw.PreferencesGroup

type structGroups struct {
	groups Groups
	page   *adw.PreferencesPage
}

func AddStructGroups(page *adw.PreferencesPage, a any) Groups {
	sg := structGroups{
		groups: make(map[string]*adw.PreferencesGroup),
		page:   page,
	}

	for i := uint(0); ; i++ {
		group := sg.page.GetGroup(i)
		if group == nil {
			break
		}
		// Prepend struct rows to existing rows
		// by removing and adding them as you cannot insert
		// by an index.
		rows := rows(group)
		for _, r := range rows {
			group.Remove(r)
		}
		sg.groups[group.GetTitle()] = group
		defer func() {
			for _, r := range rows {
				group.Add(r)
			}
		}()
	}

	v := reflect.ValueOf(a).Elem()
	for _, sf := range reflect.VisibleFields(v.Type()) {
		v := v.FieldByIndex(sf.Index)
		sg.add(v, sf)
	}

	return sg.groups
}

func groups(page *adw.PreferencesPage) (s []*adw.PreferencesGroup) {
	for i := uint(0); ; i++ {
		group := page.GetGroup(i)
		if group == nil {
			break
		}
		s = append(s, group)
	}
	return
}

func rows(group *adw.PreferencesGroup) (s []*gtk.Widget) {
	for i := uint(0); ; i++ {
		w := group.GetRow(i)
		if w == nil {
			break
		}
		s = append(s, w)
	}
	return
}

func (p *structGroups) add(v reflect.Value, sf reflect.StructField) {
	if v.Kind() == reflect.Struct {
		for _, vsf := range reflect.VisibleFields(v.Type()) {
			vv := v.FieldByIndex(vsf.Index)
			p.add(vv, vsf)
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

	displayGroup := groupName
	if displayGroup != "" {
		displayGroup = L(groupName)
	}

	if v.Kind() == reflect.Map {
		group := newMapGroup(v)
		group.SetTitle(displayGroup)
		p.page.Add(group)
		return
	}
	group, ok := p.groups[groupName]
	if !ok {
		group = adw.NewPreferencesGroup()
		p.groups[groupName] = group
		group.SetTitle(displayGroup)
		p.page.Add(group)
	}

	title := sf.Tag.Get("title")
	if title == "" {
		title = sf.Name
	}

	if title != "" {
		title = L(title)
	}

	fields := strings.Split(sf.Tag.Get("row"), ",")
	description := fields[0]
	option := ""
	if len(fields) > 1 {
		option = fields[1]
	}

	if description != "" {
		description = L(description)
	}
	if p, ok := reflect.TypeAssert[PathSelector](v); ok {
		path := newPathRow(v, p.Default())
		path.SetTitle(description)
		group.Add(&path.Widget)
	} else if t, ok := reflect.TypeAssert[Defaulter](v); ok {
		if option != "" {
			option = L(option)
		}
		opt := newOptionEntryRow(v, option, t.Default())
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
