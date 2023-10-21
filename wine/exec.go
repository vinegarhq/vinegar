package wine

import (
	"errors"
	"io"
	"log"
	"os/exec"
)

type Cmd struct {
	*exec.Cmd
}

func (p *Prefix) Command(name string, arg ...string) *Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Env = append(cmd.Environ(),
		"WINEPREFIX="+p.dir,
	)

	return &Cmd{cmd}
}

func (c *Cmd) SetOutput(w io.Writer) error {
	if c.Process != nil {
		return errors.New("SetOutput after process started")
	}
	c.Stderr = w
	c.Stdout = w
	return nil
}

func (c *Cmd) OutputPipe() (io.Reader, error) {
	if err := c.SetOutput(nil); err != nil {
		return nil, err
	}

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
