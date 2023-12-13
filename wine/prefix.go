// Package wine implements wine program command routines for
// interacting with a wineprefix [Prefix]
package wine

import (
	"io"
	"log"
)

// Prefix is a representation of a wineprefix, which is where
// WINE stores its data, which is equivalent to WINE's C:\ drive.
type Prefix struct {
	// Output specifies the descendant prefix commmand's
	// Stderr and Stdout together.
	//
	// Wine will always write to stderr instead of stdout,
	// Stdout is combined just to make that certain.
	Output io.Writer

	dir string
}

// New returns a new Prefix.
func New(dir string, out io.Writer) Prefix {
	return Prefix{
		Output: out,
		dir:    dir,
	}
}

// Dir retrieves the Prefix's directory.
func (p *Prefix) Dir() string {
	return p.dir
}

// Wine makes a new Cmd with wine as the named program.
func (p *Prefix) Wine(exe string, arg ...string) *Cmd {
	arg = append([]string{exe}, arg...)

	return p.Command("wine", arg...)
}

// Kill runs Command with 'wineserver -k' as the named program.
func (p *Prefix) Kill() {
	log.Println("Killing wineprefix")

	_ = p.Command("wineserver", "-k").Run()
}
