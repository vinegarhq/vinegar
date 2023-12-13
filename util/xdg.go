package util

import (
	"os"
	"os/exec"
)

// XDGOpen makes a *exec.Cmd with xdg-open as the named program.
func XDGOpen(file string) *exec.Cmd {
	cmd := exec.Command("xdg-open", file)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd
}
