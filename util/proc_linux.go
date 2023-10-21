//go:build linux

package util

import (
	"os"
	"path/filepath"
	"strings"
)

func CommFound(query string) bool {
	comms, _ := filepath.Glob("/proc/*/comm")

	for _, comm := range comms {
		c, err := os.ReadFile(comm)
		// The 'comm' file contains a new line, we remove it, as it will mess up
		// the query. hence 'minus'ing the length by 1 removes the newline.
		if err == nil && strings.Contains(string(c)[:len(c)-1], query) {
			return true
		}
	}

	return false
}
