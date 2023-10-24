package config

import (
	"os"
)

type Environment map[string]string

// Set will only set the given environment key and value if it isn't already set
// within the Environment.
func (e *Environment) Set(key, value string) {
	env := *e

	if _, ok := env[key]; ok {
		return
	}
	env[key] = value
}

func (e *Environment) Setenv() {
	for name, value := range *e {
		os.Setenv(name, value)
	}
}
