package main

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/sewnie/rbxweb"
)

// from version-fa681ec445d7437c/ApplicationConfig/OAuth2Config.json
var oauthClientID = rbxweb.OAuthClientID("7968549422692352298")

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
	credentials := []struct {
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
		if err := l.addWineCred(k, keyPrefix+c.name, []byte(c.blob)); err != nil {
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
