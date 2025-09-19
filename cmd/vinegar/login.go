package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/jwijenbergh/puregotk/v4/adw"
	"github.com/jwijenbergh/puregotk/v4/glib"
	"github.com/jwijenbergh/puregotk/v4/gtk"
	"github.com/sewnie/rbxweb"
)

type login struct {
	*app

	builder *gtk.Builder
	dialog  adw.Dialog

	token *rbxweb.Token
	label gtk.Label
}

func (l *login) message(msg string) {
	slog.Info(msg)
	uiThread(func() { l.label.SetLabel(msg) })
}

func (l *login) quickLoginLoop() (*rbxweb.Login, error) {
	var code, status gtk.Label
	l.builder.GetObject("label-quick-login-code").Cast(&code)
	l.builder.GetObject("label-quick-login-status").Cast(&status)

	defer uiThread(func() {
		code.SetLabel("")
		status.SetLabel("")
		l.token = nil
	})

	for {
		if l.token == nil {
			uiThread(func() {
				status.SetLabel("Creating...")
			})

			t, err := l.rbx.AuthTokenV1.CreateToken()
			if err != nil {
				return nil, fmt.Errorf("create: %w", err)
			}
			l.token = t
			slog.Info("Created token", "token", t.Code, "expires", t.ExpirationTime)
		}

		uiThread(func() {
			code.SetLabel(l.token.Code)
			status.SetLabel(l.token.Status)
		})

		time.Sleep(1 * time.Second)

		s, err := l.rbx.AuthTokenV1.GetTokenStatus(l.token)
		if err != nil {
			if err.Error() == "CodeInvalid" {
				l.token = nil
				continue
			}
			return nil, fmt.Errorf("status: %w", err)
		}
		slog.Info("Quick Login Status",
			"status", s.Status, "account", s.AccountName)

		if s.Status == "Validated" {
			break
		}

		if s.Status == "Cancelled" {
			return nil, nil
		}
		l.token.Status = s.Status
	}

	return l.rbx.AuthV2.CreateLogin(l.token.Code, l.token.PrivateKey, rbxweb.LoginTypeToken)
}

func (ui *app) newLogin() *login {
	l := login{
		app:     ui,
		builder: gtk.NewBuilderFromResource("/org/vinegarhq/Vinegar/ui/login.ui"),
	}

	l.builder.GetObject("dialog-login").Cast(&l.dialog)
	l.builder.GetObject("label-loading").Cast(&l.label)
	l.dialog.Present(ui.GetActiveWindow().GetChild())
	closeCb := func(_ adw.Dialog) {
		l.builder.Unref()
	}
	l.dialog.ConnectClosed(&closeCb)

	var view adw.NavigationView
	l.builder.GetObject("navview").Cast(&view)

	// Rather than setting up a Action and making a new thread from there,
	// re-use the existing thread that this would be called from
	setSecurityFn := func(success *rbxweb.Login) {
		uiThread(func() { view.PushByTag("nav-page-loading") })
		defer uiThread(func() { view.Pop() })
		if err := l.setSecurity(); err != nil {
			uiThread(func() { ui.showError(err) })
		}
		uiThread(func() {
			l.dialog.Close()
			l.ActivateAction("control-toast",
				glib.NewVariantString("Logged in as "+success.User.Name))
		})
	}

	var quickLoginPage adw.NavigationPage
	l.builder.GetObject("page-quick-login").Cast(&quickLoginPage)

	var quickLoginTf glib.ThreadFunc = func(uintptr) uintptr {
		slog.Info("Quick Login sequence started")
		defer slog.Info("Quick Login sequence ended")

		success, err := l.quickLoginLoop()
		if err != nil {
			uiThread(func() { ui.showError(err) })
		} else if success != nil {
			uiThread(func() { view.Pop() }) // Return to login
			setSecurityFn(success)
		}
		return 0
	}
	showingCb := func(_ adw.NavigationPage) {
		glib.NewThread("quick-login-loop", &quickLoginTf, 0)
	}
	quickLoginPage.ConnectShowing(&showingCb)

	hiddenCb := func(_ adw.NavigationPage) {
		// Won't be called on success since token is nulled
		// on quick login loop exit
		if t := l.token; t != nil {
			slog.Info("Revoking Quick Login token")
			_ = l.rbx.AuthTokenV1.CancelToken(t)
		}
	}
	quickLoginPage.ConnectHidden(&hiddenCb)

	var loginButton adw.ButtonRow
	l.builder.GetObject("button-login").Cast(&loginButton)

	var username, password adw.EntryRow
	l.builder.GetObject("entry-username").Cast(&username)
	l.builder.GetObject("entry-password").Cast(&password)
	var usernameTf glib.ThreadFunc = func(uintptr) uintptr {
		// Don't make the user press it again :D
		uiThread(func() {
			loginButton.SetActivatable(false)
			loginButton.AddCssClass("dimmed")
		})
		defer uiThread(func() {
			loginButton.SetActivatable(true)
			loginButton.RemoveCssClass("dimmed")
		})

		slog.Info("Attempting login with username", "username", username.GetText())
		success, err := l.rbx.AuthV2.CreateLogin(username.GetText(), password.GetText(),
			rbxweb.LoginTypeUsername)
		if err != nil {
			uiThread(func() { l.showError(err) })
		} else {
			setSecurityFn(success)
		}
		return 0
	}
	act := func(_ adw.ButtonRow) {
		glib.NewThread("login-attempt", &usernameTf, 0)
	}
	loginButton.ConnectActivated(&act)

	return &l
}
