package bootstrapper

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func extract(source string, dir string) error {
	zip, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer zip.Close()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	for _, file := range zip.File {
		filePath := filepath.Join(dir, strings.ReplaceAll(file.Name, `\`, "/"))

		// ignore the destination directory, it was already created above
		if filePath == dir {
			continue
		}

		if !strings.HasPrefix(filePath, filepath.Clean(dir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", filePath)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, file.Mode()); err != nil {
				return err
			}

			continue
		}

		err := func() error {
			dest, err := os.OpenFile(dir, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				return err
			}
			defer dest.Close()
		
			zipped, err := file.Open()
			if err != nil {
				return err
			}
			defer zipped.Close()
		
			if _, err := io.Copy(dest, zipped); err != nil {
				return err
			}
		
			return nil
		}()

		if err != nil {
			return err
		}
	}

	return nil
}
