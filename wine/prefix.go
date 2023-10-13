package wine

import (
	"io"
	"log"
	"os"
)

type Prefix struct {
	Dir    string
	Output io.Writer
}

func New(dir string) Prefix {
	return Prefix{
		Dir:    dir,
		Output: os.Stderr,
	}
}

func (p *Prefix) Wine(exe string, arg ...string) *Cmd {
	arg = append([]string{exe}, arg...)

	return p.Command("wine", arg...)
}

func (p *Prefix) Kill() {
	log.Println("Killing wineprefix")

	_ = p.Command("wineserver", "-k").Run()
}
