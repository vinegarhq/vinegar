package main

import (
	"os"
	"runtime/debug"
)

var Version string

func main() {
	debug.SetPanicOnFault(true)

	ui := New()
	defer ui.Unref()

	if code := ui.app.Run(len(os.Args), os.Args); code > 0 {
		os.Exit(code)
	}
}
