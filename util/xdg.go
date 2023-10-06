package util

import (
	"os"
	"os/exec"
)

func XDGOpen(file string) *exec.Cmd {
	cmd := exec.Command("xdg-open", file)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd
}
