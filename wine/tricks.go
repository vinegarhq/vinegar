package wine

import (
	"log"
	"os/exec"
	"path/filepath"

	"github.com/vinegarhq/vinegar/internal/config"
)

func isWinerootTricksAvaliable(wineroot string) (bool, error) {
	_, err := exec.LookPath(filepath.Join(wineroot, "bin", "winetricks"))
	if err != nil {
		return false, err
	}
	return true, nil

}

func (p *Prefix) DisableCrashDialogs() error {
	log.Println("Disabling Crash dialogs")

	return p.RegistryAdd("HKEY_CURRENT_USER\\Software\\Wine\\WineDbg", "ShowCrashDialog", REG_DWORD, "")
}

func (p *Prefix) LaunchWinetricks(cfg *config.Config) error {
	log.Println("Launching winetricks")
	result, err := isWinerootTricksAvaliable(cfg.WineRoot)
	if err != nil && !result {
		log.Println("winetricks not found in wineroot: " + err.Error() + ", falling back to system winetricks")
		return p.Command("winetricks").Run()
	}

	return p.Command(filepath.Join(cfg.WineRoot, "bin", "winetricks")).Run()
}
