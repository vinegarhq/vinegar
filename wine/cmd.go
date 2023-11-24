package wine

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
)

type Cmd struct {
	*exec.Cmd

	// in-order to ensure that the WINEPREFIX environment
	// variable cannot be tampered with.
	prefixDir string
}

// Command returns a passthrough Cmd struct to execute the named
// program with the given arguments.
//
// The command's Stderr and Stdout will be set to their os counterparts
// if the prefix's Output is nil.
func (p *Prefix) Command(name string, arg ...string) *Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if p.Output != nil {
		cmd.Stdout = p.Output
		cmd.Stderr = p.Output
	}

	return &Cmd{
		Cmd:       cmd,
		prefixDir: p.dir,
	}
}

// SetOutput set's the command's standard output and error to
// the given io.Writer.
func (c *Cmd) SetOutput(o io.Writer) {
	c.Stdout = o
	c.Stderr = o
}

// OutputPipe erturns a pipe that will be a MultiReader
// of StderrPipe and StdoutPipe, it will set both Stdout
// and Stderr to nil once ran.
func (c *Cmd) OutputPipe() (io.Reader, error) {
	if c.Process != nil {
		return nil, errors.New("OutputPipe after process started")
	}

	c.SetOutput(nil)

	e, err := c.StderrPipe()
	if err != nil {
		return nil, err
	}

	o, err := c.StdoutPipe()
	if err != nil {
		return nil, err
	}

	return io.MultiReader(e, o), nil
}

func (c *Cmd) Start() error {
	c.Env = append(c.Environ(),
		"WINEPREFIX="+c.prefixDir,
	)

	log.Printf("Starting command: %s", c.String())

	return c.Cmd.Start()
}

func (c *Cmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}
