package config

import (
	"bytes"
	"testing"
)

func TestConfigEncode(t *testing.T) {
	cfg := Default()
	buf := new(bytes.Buffer)

	if err := cfg.Encode(buf); err != nil {
		t.Error(err)
	}

	if buf.Len() != 0 {
		t.Error("output diff should be empty")
	}

	cfg.Debug = true
	cfg.Env["BAR"] = "1"
	cfg.Studio.GameMode = false
	cfg.Studio.WebView = ""
	cfg.Studio.FFlags["DFIntFoo"] = 960
	cfg.Studio.Env["FOO"] = "1"
	delete(cfg.Env, "WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS") // It's only temporary.

	if err := cfg.Encode(buf); err != nil {
		t.Error(err)
	}

	exp := `debug = true

[env]
BAR = "1"

[studio]
gamemode = false
webview = ""
[studio.env]
FOO = "1"
[studio.fflags]
DFIntFoo = 960
`

	if buf.String() != exp {
		t.Errorf("expected diff, got %v", buf.String())
	}
}
