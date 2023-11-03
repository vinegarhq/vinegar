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
