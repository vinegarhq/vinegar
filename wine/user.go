package wine

import (
	"os/user"
	"path/filepath"
)

func (p *Prefix) AppDataDir() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	return filepath.Join(p.Dir, "drive_c", "users", user.Username, "AppData"), nil
}
