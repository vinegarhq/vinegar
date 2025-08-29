package main

import (
	"crypto/rand"
	"crypto/rc4"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"
)

var (
	credPath = `HKEY_CURRENT_USER\Software\Wine\Credential Manager`
	// UserID comes afterwards
	authPrefix = credPath + `\Generic: https://www.roblox.com:RobloxStudioAuth`
	secPrefix  = authPrefix + `.ROBLOSECURITY`
	userPrefix = authPrefix + `userid`
)

func (a *app) getSecurity() error {
	slog.Info("Retrieving logged in user authentication")

	keys, err := a.pfx.RegistryQuery(credPath, ``)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	// Encryption key will be unavailable, no authentication
	// was performed yet
	if len(keys) == 0 {
		return errors.New("user not logged in")
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

	return fmt.Errorf("user %s cookie not found", user)
}

// workaround rc4.Cipher KSA to keep the original key intact
func keyStream(c *rc4.Cipher, subKey []byte) string {
	cpy := *c
	sec := make([]byte, len(subKey))
	cpy.XORKeyStream(sec, subKey)
	return string(sec)
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

func filetime(t time.Time) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(
		t.UTC().UnixNano()/100+116444736000000000))
	return b
}
