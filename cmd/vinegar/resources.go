package main

import (
	_ "embed"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
)

//go:embed vinegar.gresource
var gResource []byte

func init() {
	b := glib.NewBytes(
		(uintptr)(unsafe.Pointer(&gResource[0])),
		uint(len(gResource)),
	)
	r, err := gio.NewResourceFromData(b)
	if err != nil {
		panic(err)
	}
	gio.ResourcesRegister(r)
}
