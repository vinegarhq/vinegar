//go:build freebsd

package sysinfo

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"fmt"
)

const drmPath = "/dev/dri"

// Determines if a Card is 'embedded' or not, by checking
// if one of these displays belong to the card.
var embeddedDisplays = []string{"eDP", "LVDS"}

func getCards() (cs []Card) {
	// Get the list of available DRM cards
	drmCards, err := filepath.Glob(path.Join(drmPath, "card[0-9]*"))
	if err != nil {
		return // Handle any potential errors gracefully
	}

	for i, c := range drmCards {
		// Constructing paths for device and driver detection
		devPath := path.Join(c, "device")

		// Get the actual driver name using sysctl
		var driver string
		driver, err = getDriverFromSysctl(i) // Pass the index to handle each card
		if err != nil {
			driver = "unknown" // Fallback if driver cannot be detected
		}

		cs = append(cs, Card{
			Index:    i,
			Path:     c,
			Device:   devPath,
			Driver:   driver,
			Embedded: isEmbedded(c),
		})
	}

	return
}

// Function to get the driver from FreeBSD's sysctl interface for each card
func getDriverFromSysctl(cardIndex int) (string, error) {
	driverSysctl := fmt.Sprintf("hw.dri.%d.name", cardIndex)
	cmd := exec.Command("sysctl", driverSysctl)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	driverName := strings.TrimSpace(string(output))
	parts := strings.SplitN(driverName, ":", 2)
	if len(parts) > 1 {
		// Extract the driver name
		driver := strings.TrimSpace(parts[1])

		return driver, nil
	}

	return "", fmt.Errorf("driver not found")
}



// Function to determine if a card is embedded
func isEmbedded(cardPath string) bool {
	// Walk through the device files to check for embedded displays
	var isEmbed bool
	filepath.Walk(drmPath, func(p string, f os.FileInfo, err error) error {
		if !strings.HasPrefix(p, cardPath) {
			return nil
		}

		// Check if the card contains an embedded display
		for _, hwd := range embeddedDisplays {
			if strings.Contains(p, hwd) {
				isEmbed = true
			}
		}

		return nil
	})

	return isEmbed
}
