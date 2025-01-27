package main

import (
	"flag"
	"os"
)

var Version string

func main() {
	flag.Parse()

	ui := New()
	defer ui.Unref()

	if code := ui.app.Run(1, []string{os.Args[0]}); code > 0 {
		os.Exit(code)
	}
}
