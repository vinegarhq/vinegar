package util

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Extract or Un-Zip a given file to a destination directory.
func Unzip(source string, destDir string) error {
	zip, err := zip.OpenReader(source)
	if err != nil {
		return err
	}

	for _, file := range zip.File {
		// Roblox's Zip Files have windows paths in them
		filePath := filepath.Join(destDir, strings.ReplaceAll(file.Name, `\`, "/"))

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, file.Mode()); err != nil {
				return err
			}

			continue
		}

		destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		fileZipped, err := file.Open()
		if err != nil {
			return err
		}

		if _, err := io.Copy(destFile, fileZipped); err != nil {
			return err
		}

		destFile.Close()
		fileZipped.Close()
	}

	zip.Close()

	return nil
}
