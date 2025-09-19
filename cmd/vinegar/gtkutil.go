package main

import (
	_ "embed"
	"iter"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
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

func idle(bg func()) {
	var idlecb glib.SourceFunc
	idlecb = func(uintptr) bool {
		defer glib.UnrefCallback(&idlecb)
		bg()
		return false
	}
	glib.IdleAdd(&idlecb, 0)
}

func arraySeq[T any](arr uintptr, n int) iter.Seq[T] {
	// https://go.dev/wiki/cgo#turning-c-arrays-into-go-slices
	slice := (*[1 << 28]T)(unsafe.Pointer(arr))[:n:n]
	return func(yield func(T) bool) {
		for _, s := range slice {
			if !yield(s) {
				return
			}
		}
	}
}

// Workaround for casting interfaces with new
// pointer breaking the underlying type
type objectPtr[T any] interface {
	GoPointer() uintptr
	SetGoPointer(uintptr)
	*T
}

func listSeq[T any, P objectPtr[T]](list *glib.List) iter.Seq[P] {
	return func(yield func(P) bool) {
		for ; list != nil; list = list.Next {
			var obj T
			p := P(&obj)
			p.SetGoPointer(list.Data)
			if !yield(p) {
				return
			}
		}
	}
}

// Some functions take in an gio.AsyncResult which gio.AsyncResultBase
// (base interface implementer, able to cast) does not implement.
// https://github.com/jwijenbergh/puregotk/issues/23
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
