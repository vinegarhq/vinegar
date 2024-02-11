package wine

import (
	"strconv"
)

// Winetricks runs winetricks within the Prefix.
func (p *Prefix) Winetricks() error {
	return p.Command("winetricks").Run()
}

// SetDPI sets the Prefix's DPI to the named DPI.
func (p *Prefix) SetDPI(dpi int) error {
	return p.RegistryAdd("HKEY_CURRENT_USER\\Control Panel\\Desktop", "LogPixels", REG_DWORD, strconv.Itoa(dpi))
}
