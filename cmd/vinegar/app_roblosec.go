package main

import (
	"crypto/rand"
	"crypto/rc4"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

var credPath = `HKEY_CURRENT_USER\Software\Wine\Credential Manager`

// TODO: use for rbxweb.Client
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

func (b *app) getWineCredKey() ([]byte, error) {
	k, err := b.pfx.RegistryQuery(credPath, `EncryptionKey`)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	if k != nil {
		return k[0].Subkeys[0].Value.([]byte), nil
	}

	// Generate one since Wine will only do so if an application uses the
	// Credential Manager API, not the registry.
	enc := make([]byte, 8)
	_, _ = rand.Read(enc[:])
	if err := b.pfx.RegistryAdd(credPath, "EncryptionKey", enc); err != nil {
		return nil, fmt.Errorf("add: %w", err)
	}
	return enc, nil
}

func (b *app) addWineCred(key []byte, name string, blob []byte) error {
	c, err := rc4.NewCipher(key)
	if err != nil {
		return fmt.Errorf("cipher: %w", err)
	}
	encr := make([]byte, len(blob))
	c.XORKeyStream(encr, blob)

	slog.Info("Writing Wine Credential", "name", name)

	for _, sk := range []struct {
		name  string
		value any
	}{
		{"", name}, // (Default)
		{"Flags", uint32(0)},
		{"LastWritten", filetime(time.Now())},
		{"Password", encr},
		{"Persist", uint32(2)},
		{"Type", uint32(1)},
	} {
		if err := b.pfx.RegistryAdd(credPath+`\Generic: `+name, sk.name, sk.value); err != nil {
			return fmt.Errorf("add %s: %w", sk.name, err)
		}
	}
	return nil
}
