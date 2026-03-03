package main

import (
	"crypto/rc4"
	"errors"
	"fmt"
	"log/slog"

	. "github.com/pojntfx/go-gettext/pkg/i18n"
	"github.com/sewnie/wine"
)

var (
	authNamePrefix = `Generic: https://www.roblox.com:RobloxStudioAuth`
	credRegPath    = `HKCU\Software\Wine\Credential Manager`
)

func (b *bootstrapper) getSecurity(offline *wine.Registry) error {
	if offline == nil {
		// Wineprefix is not initialized
		return nil
	}

	defer b.performing()()
	b.message(L("Acquiring user authentication"))

	cred := offline.Query(credRegPath)
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

	sec := cred.Query(authNamePrefix + `.ROBLOSECURITY` + user)
	if sec == nil {
		slog.Warn("ROBLOSECURITY cookie not found", "user", user)
		return nil
	}
	b.rbx.Security = keyStream(c, sec.GetValue("Password").Data.([]byte))
	return nil
}

// workaround rc4.Cipher KSA to keep the original key intact
func keyStream(c *rc4.Cipher, subKey []byte) string {
	cpy := *c
	sec := make([]byte, len(subKey))
	cpy.XORKeyStream(sec, subKey)
	return string(sec)
}
