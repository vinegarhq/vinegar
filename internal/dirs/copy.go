package dirs

import (
	"errors"
	"bytes"
	"os"

	"log/slog"
	"path/filepath"

	cp "github.com/otiai10/copy"
)

func CompareLink(src, dest string) (bool, error) {
	info, err := os.Stat(src)
	if err != nil {
		return false, err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		srcSymPath, err := filepath.EvalSymlinks(src)
		if err != nil {
			return false, err
		}

		destSymPath, err := filepath.EvalSymlinks(dest)
		if err != nil {
			return false, err
		}

		if destSymPath != srcSymPath {
			return false, nil
		}
	}

	return true, nil
}

func OverlayDir(overlay, dir string) error {
	_, err := os.Stat(overlay)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}

	slog.Info("Copying Overlay directory's files", "dir", dir)

	return cp.Copy(overlay, dir, cp.Options{
		Sync: true,
		PreserveOwner: true,
		PreserveTimes: true,
		Skip: func(info os.FileInfo, src, dest string) (bool, error) {
			if info.IsDir() {
				return false, nil
			}

			if res, err := CompareLink(src, dest); err == nil && res {
				return true, nil
			}

			destContent, err := os.ReadFile(dest)
			if err != nil {
				return false, nil
			}

			srcContent, err := os.ReadFile(src)
			if err != nil {
				return false, nil
			}

			if !bytes.Equal(srcContent, destContent) {
				err = os.Remove(dest)
				if err != nil {
					return false, err
				}
			}

			return false, nil
		},
	})
}
