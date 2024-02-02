package dirs

import (
	"errors"
	"log/slog"
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

	slog.Info("Copying Overlay directory's files", "dir", dir)

	return cp.Copy(Overlay, dir)
}
