package dirs

import (
	"errors"
	"log"
	"os"

	cp "github.com/otiai10/copy"
)

func OverlayDir(dir string) error {
	_, err := os.Stat(Overlay)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	log.Println("Copying Overlay directory's files")

	return cp.Copy(Overlay, dir)
}
