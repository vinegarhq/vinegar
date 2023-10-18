package sysinfo

import (
	"os"
	"bufio"
	"strings"
)

type Distro struct {
	Name string
	Version string
}

func GetDistro() (Distro, error) {
	var d Distro

	f, err := os.Open("/etc/os-release")
	if err != nil {
		return Distro{}, err
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
	if err := s.Err(); err != nil {
		return Distro{}, err
	}
	
	return d, nil
}

func (d Distro) String() string {
	if d.Name == "" {
		d.Name = "Linux"
	}

	if d.Version == "" {
		d.Version = "Linux"
	}

	return d.Name + " " + d.Version
}
