package config

import (
	"os"
)

// Environment is a map representation of a operating environment
// with it's variables.
type Environment map[string]string

// Set will only set the given environment key and value
// if it isn't already set within Environment.
func (e Environment) Set(key, value string) {
	if _, ok := e[key]; ok {
		return
	}
	e[key] = value
}

// Setenv will apply the environment's variables onto the
// global environment using os.Setenv.
func (e Environment) Setenv() {
	for name, value := range e {
		os.Setenv(name, value)
	}
}
