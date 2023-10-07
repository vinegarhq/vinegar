package bootstrapper

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func extract(src string, dir string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	for _, f := range r.File {
		dest := filepath.Join(dir, strings.ReplaceAll(f.Name, `\`, "/"))

		// ignore the destination directory, it was already created above
		if dir == dest {
			continue
		}

		if !strings.HasPrefix(dest, filepath.Clean(dir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", dest)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(dest, f.Mode()); err != nil {
				return err
			}

			continue
		}

		if err := extractFile(f, dest); err != nil {
			return err
		}
	}

	return nil
}

func extractFile(src *zip.File, dest string) error {
	f, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, src.Mode())
	if err != nil {
		return err
	}
	defer f.Close()

	z, err := src.Open()
	if err != nil {
		return err
	}
	defer z.Close()

	if _, err := io.Copy(f, z); err != nil {
		return err
	}

	return nil
}
