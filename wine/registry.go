package wine

import (
	"errors"
)

type RegistryType string

const (
	REG_SZ        RegistryType = "REG_SZ"
	REG_MULTI_SZ  RegistryType = "REG_MULTI_SZ"
	REG_EXPAND_SZ RegistryType = "REG_EXPAND_SZ"
	REG_DWORD     RegistryType = "REG_DWORD"
	REG_QWORD     RegistryType = "REG_QWORD"
	REG_BINARY    RegistryType = "REG_BINARY"
	REG_NONE      RegistryType = "REG_NONE"
)

func (p *Prefix) RegistryAdd(key, value string, rtype RegistryType, data string) error {
	if key == "" {
		return errors.New("no registry key given")
	}

	return p.Wine("reg", "add", key, "/v", value, "/t", string(rtype), "/d", data, "/f").Run()
}
