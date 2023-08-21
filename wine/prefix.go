package wine

import (
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

type Prefix struct {
	Dir      string
	Launcher []string
	Output   io.Writer
	Version  string
}

func New(dir string, ver string) Prefix {
	if ver == "" {
		ver = "win10"
	}

	return Prefix{
		Dir:     dir,
		Output:  os.Stderr,
		Version: ver,
	}
}

func (p *Prefix) ExecWine(args ...string) error {
	args = append([]string{"wine"}, args...)

	if len(p.Launcher) > 0 {
		log.Printf("Using launcher: %s", p.Launcher)
		args = append(p.Launcher, args...)
	}

	return p.Exec(args[0], args[1:]...)
}

func (p *Prefix) Exec(name string, args ...string) error {
	log.Printf("Executing: %s %s", name, args)

	cmd := exec.Command(name, args...)
	cmd.Stderr = p.Output
	cmd.Stdout = p.Output
	cmd.Env = append(cmd.Environ(),
		"WINEPREFIX="+p.Dir,
	)

	return cmd.Run()
}

func (p *Prefix) Setup() error {
	if _, err := os.Stat(filepath.Join(p.Dir, "drive_c", "windows")); err == nil {
		return nil
	}

	return p.Initialize()
}

func (p *Prefix) Initialize() error {
	log.Println("Initializing wineprefix")

	if err := os.MkdirAll(p.Dir, 0o755); err != nil {
		return err
	}

	if err := p.Exec("wineboot", "-i"); err != nil {
		return err
	}

	log.Println("Setting wineprefix version to", p.Version)

	return p.Exec("winecfg", "/v", p.Version)
}

func (p *Prefix) Kill() {
	log.Println("Killing wineprefix")

	_ = p.Exec("wineserver", "-k")
}

func (p *Prefix) Interrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-c
		p.Kill()
	}()
}
