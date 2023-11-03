package bootstrapper

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/vinegarhq/vinegar/roblox/version"
	"github.com/vinegarhq/vinegar/util"
)

type File struct {
	Path     string
	Checksum string
}

type Manifest []File

var ErrInvalidManifest = errors.New("invalid manifest given")

func FetchManifest(ver *version.Version) (Manifest, error) {
	cdn, err := CDN()
	if err != nil {
		return Manifest{}, err
	}
	durl := cdn + channelPath(ver.Channel) + ver.GUID
	url := durl + "-rbxManifest.txt"

	log.Printf("Fetching files manifest for %s (%s)", ver.GUID, url)

	smanif, err := util.Body(url)
	if err != nil {
		return Manifest{}, fmt.Errorf("fetch %s manifest: %w", ver.GUID, err)
	}

	// Because the manifest ends with also a newline, it has to be removed.
	manif := strings.Split(smanif, "\r\n")
	if len(manif) > 0 && manif[len(manif)-1] == "" {
		manif = manif[:len(manif)-1]
	}

	return ParseFiles(manif)
}

func ParseFiles(manifest []string) (Manifest, error) {
	var m Manifest

	if (len(manifest) % 2) != 0 {
		return m, ErrInvalidManifest
	}

	for i := 0; i <= len(manifest)-2; i += 2 {
		if manifest[i] == "MicrosoftEdgeWebview2Setup.exe" {
			continue
		}

		m = append(m, File{
			Path:     strings.ReplaceAll(manifest[i], `\`, "/"),
			Checksum: manifest[i+1],
		})
	}

	return m, nil
}
