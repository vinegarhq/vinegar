package wine

import (
	"log"
	"strconv"
)

func (p *Prefix) Winetricks() error {
	log.Println("Launching winetricks")

	return p.Command("winetricks").Run()
}

func (p *Prefix) SetDPI(dpi int) error {
	return p.RegistryAdd("HKEY_CURRENT_USER\\Control Panel\\Desktop", "LogPixels", REG_DWORD, strconv.Itoa(dpi))
}
