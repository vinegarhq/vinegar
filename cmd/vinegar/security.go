package main

import (
	"crypto/rc4"
	"errors"
	"strings"
)

var credPath = `HKEY_CURRENT_USER\Software\Wine\Credential Manager`

func (a *app) getSecurity() (string, error) {
	keys, err := a.pfx.RegistryQuery(credPath, ``)
	if err != nil {
		return "", err
	}

	var c *rc4.Cipher
	for _, k := range keys {
		if k.Key != credPath && !strings.Contains(k.Key, "RobloxStudioAuth") {
			continue
		}
		// following subkeys usually only belong to the ROBLOSECURITY or root
		// credential registry key
		for _, sk := range k.Subkeys {
			switch sk.Name {
			case "EncryptionKey": // [Software\Wine\Credential Manager
				c, err = rc4.NewCipher(sk.Value.([]byte))
				if err != nil {
					return "", err
				}
			case "Password": // Software\Wine\Credential Manager\Generic: https://www.roblox.com:RobloxStudioAuth.
				sec := make([]byte, len(sk.Value.([]byte)))
				c.XORKeyStream(sec, sk.Value.([]byte))
				return string(sec), nil
			}
		}
	}

	return "", errors.New("failed to locate security")
}
