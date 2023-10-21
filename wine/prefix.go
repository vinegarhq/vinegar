package wine

import (
	"log"
)

type Prefix struct {
	dir string
}

func New(dir string) Prefix {
	return Prefix{
		dir: dir,
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
