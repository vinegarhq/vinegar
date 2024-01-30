package config

import (
	"os"
	"testing"
)

func TestEnv(t *testing.T) {
	e := Environment{
		"MEOW": "purr",
	}

	if e.Set("MEOW", "miaow"); e["MEOW"] != "purr" {
		t.Fatal("expected Set not override")
	}

	if e.Setenv(); os.Getenv("MEOW") != "purr" {
		t.Fatal("expected Setenv set global environment")
	}
}

func TestSanitizeEnv(t *testing.T) {
	AllowedEnv = []string{"ALLOWED"}
	e := Environment{
		"ALLOWED":  "im not impostor",
		"IMPOSTOR": "im impostor",
	}

	e.Setenv()
	SanitizeEnv()

	if os.Getenv("ALLOWED") != e["ALLOWED"] {
		t.Fatal("want allowed var, got sanitized")
	}

	if os.Getenv("IMPOSTOR") != "" {
		t.Fatal("want sanitized impostor var, got value")
	}
}
