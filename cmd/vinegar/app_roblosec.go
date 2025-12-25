package main

import (
	"crypto/rc4"
	"errors"
	"fmt"
	"log/slog"
	"slices"
)

var (
	credPath = `HKEY_CURRENT_USER\Software\Wine\Credential Manager`
	// UserID comes afterwards
	authPrefix = credPath + `\Generic: https://www.roblox.com:RobloxStudioAuth`
	secPrefix  = authPrefix + `.ROBLOSECURITY`
	userPrefix = authPrefix + `userid`
)

func (a *app) getSecurity() error {
	keys, err := a.pfx.RegistryQuery(credPath, ``)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(keys) == 0 {
		return errors.New("no authentication")
	}

	// Credential Manager first root subkey
	c, err := rc4.NewCipher(keys[0].Subkeys[0].Value.([]byte))
	if err != nil {
		return fmt.Errorf("cipher: %w", err)
	}

	var user string
	// The current user comes last. Read in reverse order to get the key
	// and get the ROBLOSECURITY entries.
	// A race condition is most likely to occur here, as it assumes Wine
	// retains and has a specific order for how keys are displayed.
	for _, k := range slices.Backward(keys) {
		switch k.Key {
		case userPrefix:
			user = keyStream(c, k.Subkeys[3].Value.([]byte))
		case secPrefix + user:
			a.rbx.Security = keyStream(c, k.Subkeys[3].Value.([]byte))
			slog.Info("Using user for authentication", "user", user)
			return nil
		}
	}

	if user == "" {
		return fmt.Errorf("not logged in")
	}

	return fmt.Errorf("user %s cookie not found", user)
}

// workaround rc4.Cipher KSA to keep the original key intact
func keyStream(c *rc4.Cipher, subKey []byte) string {
	cpy := *c
	sec := make([]byte, len(subKey))
	cpy.XORKeyStream(sec, subKey)
	return string(sec)
}
