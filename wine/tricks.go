package wine

import (
	"log"
	"strconv"
)

// Winetricks runs Command with winetricks as the named program
func (p *Prefix) Winetricks() error {
	log.Println("Launching winetricks")

	return p.Command("winetricks").Run()
}

// SetDPI calls RegistryAdd with the intent to set the DPI to the named dpi
func (p *Prefix) SetDPI(dpi int) error {
	return p.RegistryAdd("HKEY_CURRENT_USER\\Control Panel\\Desktop", "LogPixels", REG_DWORD, strconv.Itoa(dpi))
}
