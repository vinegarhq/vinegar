package wine

import (
	"os/user"
	"path/filepath"
)

// AppDataDir gets the current user and retrieves the user's AppData
// in the Prefix's users directory
func (p *Prefix) AppDataDir() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	return filepath.Join(p.dir, "drive_c", "users", user.Username, "AppData"), nil
}
