package gtkutil

import (
	_ "embed"
	"iter"
	"unsafe"

	"github.com/jwijenbergh/puregotk/v4/glib"
)

// Workaround for casting interfaces with new
// pointer breaking the underlying type
type objectPtr[T any] interface {
	GoPointer() uintptr
	SetGoPointer(uintptr)
	*T
}

func Array[T any](arr uintptr, n int) iter.Seq[T] {
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

func List[T any, P objectPtr[T]](list *glib.List) iter.Seq[P] {
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
