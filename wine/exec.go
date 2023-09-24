package wine

import (
	"log"
	"os/exec"
)

type Cmd struct {
	*exec.Cmd
}

func (p *Prefix) Command(name string, arg ...string) *Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = p.Output
	cmd.Stdout = p.Output
	cmd.Env = append(cmd.Environ(),
		"WINEPREFIX="+p.Dir,
	)

	return &Cmd{cmd}
}

func (c *Cmd) Start() error {
	log.Println("Starting command in background: %s", c.String())
	
	return c.Cmd.Start()
}

func (c *Cmd) Run() error {
	log.Printf("Running command: %s", c.String())
	
	return c.Cmd.Run()
}
