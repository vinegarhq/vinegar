package gutil

import (
	_ "embed"
	"iter"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gobject"
	"github.com/jwijenbergh/puregotk/v4/gtk"
)

// Object represents any type that implements both the base
// [gobject.Object] functions and [gobject.Ptr].
type Object[T any] interface {
	GoObject[T]

	Cast(v gobject.Ptr)
	ConnectSignal(signal string, cb *func()) uint32
	DisconnectSignal(handler uint32)
}

// GoObject is an interface that represents any type that
// implements [gobject.Ptr], a possible C structure.
type GoObject[T any] interface {
	gobject.Ptr
	*T
}

// ConnectBuilderSimple is a small wrapper for [ConnectBuilder] to
// create a callback without needing to remember the original
// object it is for, useful for a [gtk.Button].
func ConnectBuilderSimple(
	builder *gtk.Builder, name, signal string, callback func(),
) {
	ConnectBuilder[gobject.Object](builder, name, signal, func(_ *gobject.Object) {
		callback()
	})
}

// ConnectBuilder is a helper wrapper for combining [GetObject]
// and [ConnectSignal], used to quickly connect a signal from a
// builder object, without needing to remember the original object
// prior to setup.
func ConnectBuilder[T any, O Object[T]](
	builder *gtk.Builder, name, signal string, callback func(O),
) {
	obj := O(new(T))
	builder.GetObject(name).Cast(obj)

	ConnectSignal(obj, signal, func() {
		callback(obj)
	})
}

// ConnectSignal handles the boilerplate of connecting a signal [signal] to the
// given pointer object. ConnectSignal may call the callback to a different
// object if the pointer to the given object changes.
func ConnectSignal[T any, O Object[T]](object O, signal string, callback func()) {
	cb := func() {
		callback()
	}
	object.ConnectSignal(signal, &cb)
}

// GetObject is a wrapper function for retrieving the object at [name]
// for the given builder, reducing boilerplate for specifying the type
// at a variable and asserting.
func GetObject[T any, P GoObject[T]](builder *gtk.Builder, name string) T {
	var obj T
	ptr := P(&obj)
	builder.GetObject(name).Cast(ptr)
	return obj
}

// Slice is an adapatation of [unsafe.Slice] to allow
// a different type from the origin array.
//
// https://go.dev/wiki/cgo#turning-c-arrays-into-go-slices
func Slice[T any](arr uintptr, size uint) []T {
	return (*[1 << 28]T)(unsafe.Pointer(arr))[:size:size]
}

// List is a helper utility to easily iterate over a [glib.List]
// and modify the values.
func List[T any, P GoObject[T]](list *glib.List) iter.Seq[P] {
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
