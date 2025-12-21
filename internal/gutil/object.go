package gutil

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

// Slice is an adapatation of [unsafe.Slice] to allow
// a different type from the origin array.
//
// https://go.dev/wiki/cgo#turning-c-arrays-into-go-slices
func Slice[T any](arr uintptr, size uint) []T {
	return (*[1 << 28]T)(unsafe.Pointer(arr))[:size:size]
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
