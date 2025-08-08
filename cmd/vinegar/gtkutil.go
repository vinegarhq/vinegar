package main

import (
	"C"
	_ "embed"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
)

// used more often than you think
var null = uintptr(unsafe.Pointer(nil))

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

func idle(bg func()) {
	var idlecb glib.SourceFunc
	idlecb = func(uintptr) bool {
		defer glib.UnrefCallback(&idlecb)
		bg()
		return false
	}
	glib.IdleAdd(&idlecb, 0)
}

// GTK library puregotk does not support slices as a transformation in
// certain types.
func cGoStringArray(p uintptr) (res []string) {
	for {
		s := (**C.char)(unsafe.Pointer(p))
		if *s == nil {
			break
		}
		res = append(res, C.GoString(*s))
		glib.Free(p)
		p += unsafe.Sizeof(uintptr(0))
	}
	return
}

// Exists due to the fact that puregotk's AsyncResult
// has no transformation FromInternalPtr function,
// which is required to be used in AdwMessageDialog.

type asyncResult struct {
	Ptr uintptr
}

func asyncResultFromInternalPtr(ptr uintptr) *asyncResult {
	return &asyncResult{
		Ptr: ptr,
	}
}

func (x *asyncResult) GoPointer() uintptr {
	return x.Ptr
}

func (x *asyncResult) SetGoPointer(ptr uintptr) {
	x.Ptr = ptr
}

func (x *asyncResult) GetSourceObject() *gobject.Object {
	cret := gio.XGAsyncResultGetSourceObject(x.GoPointer())
	if cret == 0 {
		return nil
	}

	return &gobject.Object{
		Ptr: cret,
	}
}

func (x *asyncResult) GetUserData() uintptr {
	return gio.XGAsyncResultGetUserData(x.GoPointer())
}

func (x *asyncResult) IsTagged(SourceTagVar uintptr) bool {
	return gio.XGAsyncResultIsTagged(x.GoPointer(), SourceTagVar)
}

func (x *asyncResult) LegacyPropagateError() bool {
	return gio.XGAsyncResultLegacyPropagateError(x.GoPointer())
}
