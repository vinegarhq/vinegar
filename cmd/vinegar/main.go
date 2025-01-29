package main

import (
	"flag"
	"os"
	"runtime/debug"
)

var Version string

func main() {
	debug.SetPanicOnFault(true)
	flag.Parse()

	ui := New()
	defer ui.Unref()

	if code := ui.app.Run(1, []string{os.Args[0]}); code > 0 {
		os.Exit(code)
	}
}
