package main

import (
	"log"
	"os"
	"path/filepath"
)

// Loop over all proc(5)fs PID directories and check if the given query (string)
// matches the file contents of with a file called 'comm', within the PID
// directory. For simplification purposes this will use a /proc/*/comm glob instead.
// Once found a 'comm' file, simply return true; return false when not found.
func CommFound(query string) bool {
	comms, err := filepath.Glob("/proc/*/comm")
	if err != nil {
		log.Fatal("failed to locate procfs commands")
	}

	for _, comm := range comms {
		c, err := os.ReadFile(comm)
		// The 'comm' file contains a new line, we remove it, as it will mess up
		// the query. hence 'minus'ing the length by 1 removes the newline.
		if err == nil && string(c)[:len(c)-1] == query {
			return true
		}
	}

	return false
}
