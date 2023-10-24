package config

import (
	"log"
	"os"
)

type Environment map[string]string

// Set will only set the given environment key and value if it isn't already set
// within the Environment.
func (e *Environment) Set(k string, v string) {
	if _, ok := e[k]; ok {
		return
	}
	e[k] = v
}

func (e *Environment) Setenv() {
	for name, value := range *e {
		os.Setenv(name, value)
	}
}
