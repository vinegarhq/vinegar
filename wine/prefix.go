package wine

import (
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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

func (p *Prefix) Setup() error {
	if _, err := os.Stat(filepath.Join(p.Dir, "drive_c", "windows")); err == nil {
		return nil
	}

	return p.Initialize()
}

func (p *Prefix) Initialize() error {
	log.Printf("Initializing wineprefix at %s", p.Dir)

	if err := os.MkdirAll(p.Dir, 0o755); err != nil {
		return err
	}

	if err := p.Command("wineboot", "-i").Run(); err != nil {
		return err
	}

	return p.DisableCrashDialogs()
}

func (p *Prefix) Kill() {
	log.Println("Killing wineprefix")

	_ = p.Command("wineserver", "-k").Run()
}

func (p *Prefix) Interrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-c
		p.Kill()
	}()
}
