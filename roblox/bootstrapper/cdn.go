package bootstrapper

import (
	"errors"
	"log"
	"net/http"
)

var (
	ErrNoCDNFound = errors.New("no accessible Roblox deploy mirror or cdn found")
	CDNs          = []string{
		"https://setup.rbxcdn.com",
		"https://s3.amazonaws.com/setup.roblox.com",
		"https://setup-ak.rbxcdn.com",
		"https://setup-hw.rbxcdn.com",
		"https://setup-cfly.rbxcdn.com", // Fastest
		"https://roblox-setup.cachefly.net",
	}
)

// CDN returns a CDN (from CDNs) that is available.
func CDN() (string, error) {
	log.Println("Finding an accessible Roblox deploy mirror")

	for _, cdn := range CDNs {
		resp, err := http.Head(cdn + "/" + "version")
		if err != nil {
			log.Printf("deploy mirror %s: %s", cdn, errors.Unwrap(err))

			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 200 {
			log.Printf("Found deploy mirror: %s", cdn)

			return cdn, nil
		}
	}

	return "", ErrNoCDNFound
}
