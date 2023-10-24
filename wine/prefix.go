package wine

import (
	"io"
	"log"
)

type Prefix struct {
	// Output specifies the descendant prefix commmand's
	// Stderr and Stdout together.
	Output io.Writer

	dir string
}

func New(dir string, out io.Writer) Prefix {
	return Prefix{
		Output: out,
		dir:    dir,
	}
}

func (p *Prefix) Dir() string {
	return p.dir
}

func (p *Prefix) Wine(exe string, arg ...string) *Cmd {
	arg = append([]string{exe}, arg...)

	return p.Command("wine", arg...)
}

func (p *Prefix) Kill() {
	log.Println("Killing wineprefix")

	_ = p.Command("wineserver", "-k").Run()
}
