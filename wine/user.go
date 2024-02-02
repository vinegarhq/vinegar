package wine

import (
	"os/user"
	"path/filepath"
)

// AppDataDir returns the current user's AppData within the Prefix.
func (p *Prefix) AppDataDir() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	return filepath.Join(p.dir, "drive_c", "users", user.Username, "AppData"), nil
}
