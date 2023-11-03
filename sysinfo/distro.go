//go:build linux

package sysinfo

import (
	"bufio"
	"os"
	"strings"
)

func getDistro() string {
	name := "Linux"
	ver := ""

	f, err := os.Open("/etc/os-release")
	if err != nil {
		return name
	}
	defer f.Close()

	s := bufio.NewScanner(f)

	for s.Scan() {
		m := strings.SplitN(s.Text(), "=", 2)
		if len(m) != 2 {
			continue
		}

		val := strings.Trim(m[1], "\"")

		switch m[0] {
		case "PRETTY_NAME":
			name = val
		case "VERSION_ID":
			ver = val
		}
	}

	return name + " " + ver
}
