package bootstrapper

import (
	"errors"
	"log"
	"net/http"
)

var (
	ErrNoCDNFound = errors.New("failed to find an accessible roblox deploy mirror")
	CDNURLs       = []string{
		"https://setup.rbxcdn.com",
		"https://s3.amazonaws.com/setup.roblox.com",
		"https://setup-ak.rbxcdn.com",
		"https://setup-hw.rbxcdn.com",
		"https://setup-cfly.rbxcdn.com",
		"https://roblox-setup.cachefly.net",
	}
)

func CDN() (string, error) {
	log.Println("Finding an accessible Roblox deploy mirror")

	for _, cdn := range CDNURLs {
		resp, err := http.Head(cdn + "/" + "version")
		if err != nil {
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
