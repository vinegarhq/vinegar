package gtkutil

import (
	_ "embed"
	"path"

	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
)

//go:embed vinegar.gresource
var gResource []byte

func init() {
	b := glib.NewBytes(
		gResource,
		uint(len(gResource)),
	)
	r, err := gio.NewResourceFromData(b)
	if err != nil {
		panic(err)
	}
	gio.ResourcesRegister(r)
}

func Resource(elems ...string) string {
	return path.Join(append([]string{"/org/vinegarhq/Vinegar"}, elems...)...)
}

func IdleAdd(bg func()) {
	var idlecb glib.SourceFunc
	idlecb = func(uintptr) bool {
		defer glib.UnrefCallback(&idlecb)
		bg()
		return false
	}
	glib.IdleAdd(&idlecb, 0)
}
