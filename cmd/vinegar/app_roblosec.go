package main

import (
	"crypto/rc4"
	"errors"
	"fmt"
	"log/slog"
)

var (
	authNamePrefix = `Generic: https://www.roblox.com:RobloxStudioAuth`
	credRegPath    = `HKCU\Software\Wine\Credential Manager`
)

func (a *app) getSecurity() error {
	cred, err := a.pfx.RegistryQuery(credRegPath)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	if cred == nil {
		return errors.New("credential manager missing")
	}

	c, err := rc4.NewCipher(
		cred.GetValue(`EncryptionKey`).Data.([]byte))
	if err != nil {
		return fmt.Errorf("cipher: %w", err)
	}

	uk := cred.Query(authNamePrefix + `userid`)
	if uk == nil {
		return errors.New("no current user")
	}
	user := keyStream(c, uk.GetValue("Password").Data.([]byte))
	slog.Info("Using user for authentication", "user", user)

	// Surely Roblox Studio would set the userid and then forget
	// to create the ROBLOSECURITY cookie for that user?
	sec := cred.Query(authNamePrefix + `.ROBLOSECURITY` + user)
	a.rbx.Security = keyStream(c, sec.GetValue("Password").Data.([]byte))
	return nil
}

// workaround rc4.Cipher KSA to keep the original key intact
func keyStream(c *rc4.Cipher, subKey []byte) string {
	cpy := *c
	sec := make([]byte, len(subKey))
	cpy.XORKeyStream(sec, subKey)
	return string(sec)
}
