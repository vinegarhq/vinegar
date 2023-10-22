package wine

import (
	"errors"
	"os"
	"io"
	"log"
	"os/exec"
)

type Cmd struct {
	*exec.Cmd
}

// Command returns a passthrough Cmd struct to execute the named
// program with the given arguments.
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

	cmd.Env = append(cmd.Environ(),
		"WINEPREFIX="+p.dir,
	)

	return &Cmd{cmd}
}

// OutputPipe erturns a pipe that will be a MultiReader
// of StderrPipe and StdoutPipe, it will set both Stdout
// and Stderr to nil once ran.
func (c *Cmd) OutputPipe() (io.Reader, error) {
	if c.Process != nil {
		return nil, errors.New("OutputPipe after process started")
	}
	
	c.Stdout = nil
	c.Stderr = nil

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
	log.Printf("Starting command: %s", c.String())

	return c.Cmd.Start()
}

func (c *Cmd) Run() error {
	log.Printf("Running command: %s", c.String())

	return c.Cmd.Run()
}
