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

	key, err := registryBytes(cred, `EncryptionKey`)
	if err != nil {
		return fmt.Errorf("encryption key: %w", err)
	}
	c, err := rc4.NewCipher(key)
	if err != nil {
		return fmt.Errorf("cipher: %w", err)
	}

	uk := cred.Query(authNamePrefix + `userid`)
	if uk == nil {
		return errors.New("no current user")
	}
	uPass, err := registryBytes(uk, "Password")
	if err != nil {
		return fmt.Errorf("user id: %w", err)
	}
	user := keyStream(c, uPass)
	slog.Info("Using user for authentication", "user", user)

	sec := cred.Query(authNamePrefix + `.ROBLOSECURITY` + user)
	if sec == nil {
		slog.Warn("ROBLOSECURITY cookie not found", "user", user)
		return nil
	}
	secPass, err := registryBytes(sec, "Password")
	if err != nil {
		return fmt.Errorf("cookie: %w", err)
	}
	b.rbx.Security = keyStream(c, secPass)
	return nil
}

// registryBytes safely retrieves the named value from key as a byte slice,
// returning an error instead of panicking if the value is missing or of an
// unexpected type. The Wine registry is external, untrusted state (it can be
// missing or corrupt), so callers must not assume a well-formed shape.
func registryBytes(key *wine.RegistryKey, name string) ([]byte, error) {
	v := key.GetValue(name)
	if v == nil {
		return nil, fmt.Errorf("value %q missing", name)
	}
	b, ok := v.Data.([]byte)
	if !ok {
		return nil, fmt.Errorf("value %q is not binary data", name)
	}
	return b, nil
}

// workaround rc4.Cipher KSA to keep the original key intact
func keyStream(c *rc4.Cipher, subKey []byte) string {
	cpy := *c
	sec := make([]byte, len(subKey))
	cpy.XORKeyStream(sec, subKey)
	return string(sec)
}
