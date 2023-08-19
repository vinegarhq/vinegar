package util

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Extract(source string, destDir string) error {
	zip, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer zip.Close()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	for _, file := range zip.File {
		filePath := filepath.Join(destDir, strings.ReplaceAll(file.Name, `\`, "/"))

		// ignore the destination directory, it was already created above
		if filePath == destDir {
			continue
		}

		if !strings.HasPrefix(filePath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", filePath)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, file.Mode()); err != nil {
				return err
			}

			continue
		}

		if err := ExtractFile(file, filePath); err != nil {
			return err
		}
	}

	return nil
}

func ExtractFile(file *zip.File, dest string) error {
	destFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer destFile.Close()

	fileZipped, err := file.Open()
	if err != nil {
		return err
	}
	defer fileZipped.Close()

	if _, err := io.Copy(destFile, fileZipped); err != nil {
		return err
	}

	return nil
}
