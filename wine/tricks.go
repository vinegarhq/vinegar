package wine

import (
	"log"
)

func (p *Prefix) DisableCrashDialogs() error {
	log.Println("Disabling Crash dialogs")

	return p.RegistryAdd("HKEY_CURRENT_USER\\Software\\Wine\\WineDbg", "ShowCrashDialog", REG_DWORD, "")
}
