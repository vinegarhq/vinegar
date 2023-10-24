package config

import (
	"log"
	"os"
)

type Environment map[string]string

func (e Environment) SetIfUndefined(k string, v string) {
	if _, ok := e[k]; ok {
		log.Printf("Warning: env var %s is already defined. Will not redefine it.", k)
		return
	}
	e[k] = v
}

func (e *Environment) Setenv() {
	for name, value := range *e {
		os.Setenv(name, value)
	}
}
