package wine

import (
	"io"
	"log"
	"os"
	"os/exec"
)

type Prefix struct {
	Dir      string
	Launcher []string
	Output   io.Writer
}

func New(dir string) Prefix {
	return Prefix{
		Dir:    dir,
		Output: os.Stderr,
	}
}

func (p *Prefix) Command(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Stderr = p.Output
	cmd.Stdout = p.Output
	cmd.Env = append(cmd.Environ(),
		"WINEPREFIX="+p.Dir,
	)

	return cmd
}

func (p *Prefix) Exec(name string, args ...string) error {
	log.Printf("Executing: %s %s", name, args)

	return p.Command(name, args...).Run()
}

func (p *Prefix) ExecWine(args ...string) error {
	args = append([]string{"wine"}, args...)

	if len(p.Launcher) > 0 {
		log.Printf("Using launcher: %s", p.Launcher)
		args = append(p.Launcher, args...)
	}

	return p.Exec(args[0], args[1:]...)
}
