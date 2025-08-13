package main

import (
	"crypto/rand"
	"crypto/rc4"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/sewnie/rbxweb"
)

var (
	credPath = `HKEY_CURRENT_USER\Software\Wine\Credential Manager`

	// from version-fa681ec445d7437c/ApplicationConfig/OAuth2Config.json
	oauthClientID = rbxweb.OAuthClientID("7968549422692352298")
)

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

func (l *login) message(msg string) {
	slog.Info(msg)
	idle(func(){l.label.SetLabel(msg)})
}

func (l *login) setSecurity() error {
	l.message("Authenticating")
	user, err := l.rbx.UsersV1.GetAuthenticated()
	if err != nil {
		return fmt.Errorf("user: %w", err)
	}
	slog.Info("Recieved authenticated user", "id", user.ID)

	l.message("Fetching authorization")
	u, err := l.rbx.OAuthV1.GetAuthStudioURL(oauthClientID, user.ID)
	if err != nil {
		return fmt.Errorf("url: %w", err)
	}
	slog.Info("Recieved Oauth authorization", "url", u)

	l.message("Fetching OAuth token")
	t, err := l.rbx.OAuthV1.AuthStudioToken(oauthClientID, u)
	if err != nil {
		return fmt.Errorf("token: %w", err)
	}
	slog.Info("Recieved OAuth Token", "type", t.TokenType)

	l.message("Retrieving credential key")
	k, err := l.getWineCredKey()
	if err != nil {
		return fmt.Errorf("cred key: %w", err)
	}

	uid := strconv.FormatInt(int64(user.ID), 10)
	expire := strconv.FormatInt(time.Now().Unix()+t.ExpiresIn, 10)

	l.message("Setting Roblox credentials")

	// TODO: .reg file representation
	keyPrefix := `https://www.roblox.com:RobloxStudioAuth`
	credentials :=  []struct{
		name string
		blob string
	}{
		{`.ROBLOSECURITY` + uid, l.rbx.Security},
		{`accessToken` + uid, t.AccessToken},
		{`expiresAtSecSinceEpoch` + uid, expire},
		{`Cookies`, ".ROBLOSECURITY;"},
		{`oauth2RefreshToken` + uid, t.RefreshToken},
		{`userid`, uid},
	} 
	for i, c := range credentials {
		l.message(fmt.Sprintf("Applying %d of %d Wine Credentials", i+1, len(credentials)))
		if err := l.addWineCred(k, keyPrefix + c.name, []byte(c.blob)); err != nil {
			return fmt.Errorf("cred %s: %w", c.name, err)
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
