package wine

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Prefix struct {
	Dir      string
	Launcher string
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

func (p *Prefix) Exec(args ...string) error {
	log.Printf("Executing wine: %s", args)

	args = append([]string{"wine"}, args...)

	if p.Launcher != "" {
		log.Printf("Using launcher: %s", p.Launcher)
		args = append([]string{p.Launcher}, args...)
	}

	cmd := exec.Command(args[0], args[1:]...)
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

	_ = p.Exec("wineboot", "-e")
}
