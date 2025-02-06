package main

import (
	"github.com/jwijenbergh/puregotk/v4/gio"
	"github.com/jwijenbergh/puregotk/v4/gobject"
)

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
