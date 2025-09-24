package main

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/sewnie/rbxweb"
)

// from version-fa681ec445d7437c/ApplicationConfig/OAuth2Config.json
var oauthClientID = rbxweb.OAuthClientID("7968549422692352298")

func (l *login) setSecurity() error {
	uiThread(func() { l.view.PushByTag("nav-page-loading") })
	defer uiThread(func() { l.view.Pop() })

	l.message("Authenticating")
	user, err := l.app.rbx.UsersV1.GetAuthenticated()
	if err != nil {
		return fmt.Errorf("user: %w", err)
	}
	slog.Info("Recieved authenticated user", "id", user.ID)

	l.message("Fetching authorization")
	u, err := l.app.rbx.OAuthV1.GetAuthStudioURL(oauthClientID, user.ID)
	if err != nil {
		return fmt.Errorf("url: %w", err)
	}
	slog.Info("Recieved Oauth authorization", "url", u)

	l.message("Fetching OAuth token")
	t, err := l.app.rbx.OAuthV1.AuthStudioToken(oauthClientID, u)
	if err != nil {
		return fmt.Errorf("token: %w", err)
	}
	slog.Info("Recieved OAuth Token", "type", t.TokenType)

	l.message("Retrieving credential key")
	k, err := l.app.getWineCredKey()
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
		{`.ROBLOSECURITY` + uid, l.app.rbx.Security},
		{`accessToken` + uid, t.AccessToken},
		{`expiresAtSecSinceEpoch` + uid, expire},
		{`Cookies`, ".ROBLOSECURITY;"},
		{`oauth2RefreshToken` + uid, t.RefreshToken},
		{`userid`, uid},
	}
	for i, c := range credentials {
		l.message(fmt.Sprintf("Applying %d of %d Wine Credentials", i+1, len(credentials)))
		if err := l.app.addWineCred(k, keyPrefix+c.name, []byte(c.blob)); err != nil {
			return fmt.Errorf("cred %s: %w", c.name, err)
		}
	}

	uiThread(func() {
		l.dialog.ForceClose()
		l.app.ActivateAction("control-toast",
			glib.NewVariantString("Logged in as "+user.Name))
	})

	return nil
}
