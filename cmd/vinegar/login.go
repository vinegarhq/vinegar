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
	app *app

	builder *gtk.Builder
	dialog  adw.Dialog

	auth  *rbxweb.Login
	token *rbxweb.Token

	label gtk.Label
	view  adw.NavigationView

	credentials        adw.PreferencesGroup
	username, password adw.EntryRow
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

			t, err := l.app.rbx.AuthTokenV1.CreateToken()
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

		s, err := l.app.rbx.AuthTokenV1.GetTokenStatus(l.token)
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

	return l.app.rbx.AuthV2.CreateLogin(l.token.Code, l.token.PrivateKey, rbxweb.LoginTypeToken)
}

func (a *app) newLogin() *login {
	l := login{
		app:     a,
		builder: gtk.NewBuilderFromResource("/org/vinegarhq/Vinegar/ui/login.ui"),
	}

	l.builder.GetObject("dialog-login").Cast(&l.dialog)
	l.builder.GetObject("label-loading").Cast(&l.label)

	l.builder.GetObject("navview").Cast(&l.view)
	l.builder.GetObject("group-credentials").Cast(&l.credentials)
	l.builder.GetObject("entry-username").Cast(&l.username)
	l.builder.GetObject("entry-password").Cast(&l.password)

	l.dialog.Present(a.GetActiveWindow().GetChild())

	closeCb := func(_ adw.Dialog) {
		l.builder.Unref()
	}
	l.dialog.ConnectClosed(&closeCb)
	l.dialog.SetCanClose(false)
	attemptCb := func(_ adw.Dialog) {
		if l.auth == nil {
			l.dialog.ForceClose()
			return
		}
		var tf glib.ThreadFunc = func(uintptr) uintptr {
			if err := l.setSecurity(); err != nil {
				uiThread(func() { l.app.showError(err) })
			}
			return 0
		}
		glib.NewThread("login-attempt", &tf, 0)
	}
	l.dialog.ConnectCloseAttempt(&attemptCb)

	var quickLoginPage adw.NavigationPage
	l.builder.GetObject("page-quick-login").Cast(&quickLoginPage)
	showingCb := func(_ adw.NavigationPage) {
		var tf glib.ThreadFunc = func(uintptr) uintptr {
			l.quickLoginThread()
			return 0
		}
		glib.NewThread("quick-login-loop", &tf, 0)
	}
	hiddenCb := func(_ adw.NavigationPage) {
		// Won't be called on success since token is nulled
		// on quick login loop exit
		if t := l.token; t != nil {
			slog.Info("Revoking Quick Login token")
			_ = l.app.rbx.AuthTokenV1.CancelToken(t)
		}
	}
	quickLoginPage.ConnectShowing(&showingCb)
	quickLoginPage.ConnectHidden(&hiddenCb)

	act := func(_ adw.EntryRow) {
		var tf glib.ThreadFunc = func(uintptr) uintptr {
			l.usernameLoginThread()
			return 0
		}
		glib.NewThread("login-attempt", &tf, 0)
	}
	l.password.ConnectEntryActivated(&act)

	return &l
}

func (l *login) usernameLoginThread() {
	// Don't make the user press it again :D
	uiThread(func() {
		l.password.SetActivatable(false)
		l.credentials.AddCssClass("dimmed")
	})
	defer uiThread(func() {
		l.password.SetActivatable(true)
		l.credentials.RemoveCssClass("dimmed")
	})

	slog.Info("Attempting login with username", "username", l.username.GetText())
	auth, err := l.app.rbx.AuthV2.CreateLogin(l.username.GetText(), l.password.GetText(),
		rbxweb.LoginTypeUsername)

	slog.Info("Recieved info", "auth", auth)
	if err != nil {
		uiThread(func() { l.app.showError(err) })
		return
	}

	if auth.Verification == nil {
		l.auth = auth
		uiThread(func() { l.dialog.Close() })
		return
	}

	var label gtk.Label
	var code adw.EntryRow
	l.builder.GetObject("label-two-step-login-type").Cast(&label)
	l.builder.GetObject("entry-code").Cast(&code)

	uiThread(func() {
		l.view.PushByTag("nav-page-two-step-login")
		label.SetText("Enter the " + string(auth.Verification.Type) + " Code")
	})
	act := func(_ adw.EntryRow) {
		code.SetActivatable(false)
		code.AddCssClass("dimmed")
		var tf glib.ThreadFunc = func(uintptr) uintptr {
			defer uiThread(func() {
				l.view.Pop()
				code.SetActivatable(true)
				code.RemoveCssClass("dimmed")
			})
			tsToken, err := l.app.rbx.VerificationServiceV1.CreateToken(auth, code.GetText())
			if err != nil {
				uiThread(func() { l.app.showError(err) })
				return 0
			}

			_, err = l.app.rbx.AuthV3.CreateTwoStepLogin(false, tsToken)
			if err != nil {
				uiThread(func() { l.app.showError(err) })
			}
			uiThread(func() {
				l.auth = auth
				l.dialog.Close()
			})
			return 0
		}
		glib.NewThread("login-attempt", &tf, 0)
	}
	code.ConnectEntryActivated(&act)
}

func (l *login) quickLoginThread() {
	slog.Info("Quick Login sequence started")
	defer slog.Info("Quick Login sequence ended")

	auth, err := l.quickLoginLoop()
	if err != nil {
		uiThread(func() { l.app.showError(err) })
	}
	if auth == nil { // user cancelled
		return
	}

	uiThread(func() {
		l.auth = auth
		l.view.Pop() // Return to login
		l.dialog.Close()
	})
}
