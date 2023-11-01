//go:build linux

package sysinfo

import (
	"bufio"
	"os"
	"strings"
)

type distro struct {
	Name    string
	Version string
}

func getDistro() (d distro) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return
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
			d.Name = val
		case "VERSION_ID":
			d.Version = val
		}
	}

	return
}

func (d distro) String() string {
	if d.Name == "" {
		d.Name = "Linux"
	}

	return d.Name + " " + d.Version
}
