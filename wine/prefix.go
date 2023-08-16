package wine

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Prefix struct {
	Dir string
	Version string
}

func New(dir string, ver string) Prefix {
	if ver == "" {
		ver = "win10"
	}

	return Prefix{
		Dir: dir,
		Version: ver,
	}
}

func (p *Prefix) Exec(args ...string) error {
	log.Printf("Executing wine: %s", args)

	cmd := exec.Command("wine", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
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
