package config

import (
	"os"
)

type Environment map[string]string

func (e *Environment) Setenv() {
	for name, value := range *e {
		os.Setenv(name, value)
	}
}
