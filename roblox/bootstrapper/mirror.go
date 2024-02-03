package bootstrapper

import (
	"errors"
	"log/slog"
	"net/http"
)

var (
	ErrNoMirrorFound = errors.New("no accessible deploy mirror found")

	// As of 2024-02-03:
	//   setup-cfly.rbxcdn.com = roblox-setup.cachefly.net
	//   setup.rbxcdn.com = setup-ns1.rbxcdn.com = setup-ak.rbxcdn.com
	//   setup-hw.rbxcdn.com = setup-ll.rbxcdn.com = does not exist
	Mirrors = []string{
		// Sorted by speed
		"https://setup.rbxcdn.com",
		"https://setup-cfly.rbxcdn.com",
		"https://s3.amazonaws.com/setup.roblox.com",
	}
)

// Mirror returns an available mirror URL from [Mirrors].
func Mirror() (string, error) {
	slog.Info("Finding an accessible deploy mirror")

	for _, m := range Mirrors {
		resp, err := http.Head(m + "/" + "version")
		if err != nil {
			slog.Error("Bad deploy mirror", "mirror", m, "error", err)

			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 200 {
			slog.Info("Found deploy mirror", "mirror", m)

			return m, nil
		}
	}

	return "", ErrNoMirrorFound
}
