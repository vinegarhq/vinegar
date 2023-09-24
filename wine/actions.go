package wine

import (
	"bytes"
	_ "embed"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func (p *Prefix) Setup() error {
	if _, err := os.Stat(filepath.Join(p.Dir, "drive_c", "windows")); err == nil {
		return nil
	}

	return p.Initialize()
}

func (p *Prefix) DisableCrashDialogs() error {
	log.Println("Disabling Crash dialogs")

	return p.RegistryAdd("HKEY_CURRENT_USER\\Software\\Wine\\WineDbg", "ShowCrashDialog", REG_DWORD, "")
}

func (p *Prefix) Initialize() error {
	log.Println("Initializing wineprefix")

	if err := os.MkdirAll(p.Dir, 0o755); err != nil {
		return err
	}

	if err := p.Exec("wineboot", "-i"); err != nil {
		return err
	}

	return p.DisableCrashDialogs()
}

func (p *Prefix) Kill() {
	log.Println("Killing wineprefix")

	_ = p.Exec("wineserver", "-k")
}

func (p *Prefix) Regedit(reg []byte) error {
	cmd := p.Command("regedit", "-")
	cmd.Stdin = bytes.NewReader(reg)

	return cmd.Run()
}

func (p *Prefix) Interrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-c
		p.Kill()
	}()
}
