// Package wine implements wine program command routines for
// interacting with a wineprefix [Prefix]
package wine

import (
	"io"
	"log"
	"os/exec"
)

// The program used for Wine.
var Wine = "wine"

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
//
// dir must be an absolute path and has correct permissions
// to modify.
func New(dir string, out io.Writer) Prefix {
	return Prefix{
		Output: out,
		dir:    dir,
	}
}

// WineLook checks for [Wine] with exec.LookPath, and returns
// true if [Wine] is present and has no problems.
func WineLook() bool {
	_, err := exec.LookPath(Wine)
	return err == nil
}

// Dir retrieves the [Prefix]'s directory on the filesystem.
func (p *Prefix) Dir() string {
	return p.dir
}

// Wine makes a new Cmd with [Wine] as the named program.
func (p *Prefix) Wine(exe string, arg ...string) *Cmd {
	arg = append([]string{exe}, arg...)

	return p.Command(Wine, arg...)
}

// Kill runs Command with 'wineserver -k' as the named program.
func (p *Prefix) Kill() {
	log.Println("Killing wineprefix")

	_ = p.Command("wineserver", "-k").Run()
}
