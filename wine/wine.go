// Package wine implements wine program command routines for
// interacting with a wineprefix [Prefix]
package wine

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	ErrWineNotFound = errors.New("wine64 not found in $PATH or wineroot")
	ErrPrefixNotAbs = errors.New("prefix directory is not an absolute path")
)

// Prefix is a representation of a wineprefix, which is where
// WINE stores its data and is equivalent to a C:\ drive.
type Prefix struct {
	// Path to a wine installation.
	Root string

	// Stdout and Stderr specify the descendant Prefix wine call's
	// standard output and error. This is mostly reserved for logging purposes.
	// By default, they will be set to their os counterparts.
	Stderr io.Writer
	Stdout io.Writer

	dir string // Path to wineprefix
}

// New returns a new Prefix.
//
// dir must be an absolute path and has correct permissions
// to modify.
func New(dir string, root string) *Prefix {
	return &Prefix{
		Root:   root,
		Stderr: os.Stderr,
		Stdout: os.Stdout,
		dir:    dir,
	}
}

// String implements the Stringer interface.
func (p Prefix) String() string {
	return p.Dir()
}

// Dir returns the directory of the Prefix.
func (p *Prefix) Dir() string {
	return p.dir
}

// Wine returns a new Cmd with the prefix's Wine as the named program.
//
// The Wine executable used is a path to the system or Prefix's Root's 'wine64'
// if present. an attempt to resolve for a [ULWGL launcher] will be made if
// it is present and necessary environment variables will be set.
//
// [ULWGL launcher]: https://github.com/Open-Wine-Components/ULWGL-launcher
func (p *Prefix) Wine(exe string, arg ...string) *Cmd {
	wine := "wine64"

	if p.Root != "" {
		ulwgl, err := exec.LookPath(filepath.Join(p.Root, "ulwgl-run"))
		if err == nil {
			os.Setenv("STORE", "none")
			wine = ulwgl
		}

		wine = filepath.Join(p.Root, "bin", "wine64")
	}

	arg = append([]string{exe}, arg...)
	cmd := p.Command(wine, arg...)

	if cmd.Err != nil && errors.Is(cmd.Err, exec.ErrNotFound) {
		cmd.Err = ErrWineNotFound
	}

	// Always ensure its created, wine will complain if the root
	// directory doesnt exist
	if err := os.MkdirAll(p.dir, 0o755); err != nil {
		cmd.Err = err
	}

	// Wine requires a absolute path for the wineprefix
	if !filepath.IsAbs(p.dir) {
		cmd.Err = ErrPrefixNotAbs
	}

	if cmd.Args[0] == "ulwgl-run" {
		cmd.Env = append(cmd.Environ(), "PROTON_VERB=runinprefix")
	}

	return cmd
}

// Kill kills the Prefix's processes.
func (p *Prefix) Kill() error {
	return p.Wine("wineboot", "-k").Run()
}

// Init preforms initialization for first Wine instance.
func (p *Prefix) Init() error {
	return p.Wine("wineboot", "-i").Run()
}

// Update updates the wineprefix directory.
func (p *Prefix) Update() error {
	return p.Wine("wineboot", "-u").Run()
}

// Version returns the wineprefix's Wine version.
func (p *Prefix) Version() string {
	cmd := p.Wine("--version")
	cmd.Stdout = nil // required for Output()
	cmd.Stderr = nil

	ver, _ := cmd.Output()
	if len(ver) < 0 {
		return "unknown"
	}

	// remove newline
	return string(ver[:len(ver)-1])
}
