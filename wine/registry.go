package wine

import (
	"errors"
)

type RegistryType string

const (
	REG_SZ        RegistryType = "REG_SZ"
	REG_MULTI_SZ               = "REG_MULTI_SZ"
	REG_EXPAND_SZ              = "REG_EXPAND_SZ"
	REG_DWORD                  = "REG_DWORD"
	REG_QWORD                  = "REG_QWORD"
	REG_BINARY                 = "REG_BINARY"
	REG_NONE                   = "REG_NONE"
)

func (p *Prefix) RegistryAdd(key, value string, kind, data string) error {
	if key == "" {
		return errors.New("no registry key given")
	}

	return p.ExecWine("reg", "add", key, "/v", value, "/t", kind, "/d", data, "/f")
}
