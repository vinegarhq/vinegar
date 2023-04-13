package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func Exec(prog string, elog bool, args ...string) error {
	if prog == "wine" {
		PfxInit()
	}

	log.Println("Executing:", prog, args)

	cmd := exec.Command(prog, args...)

	cmd.Dir = Dirs.Cwd
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if elog {
		logFile := LogFile("exec")
		log.Println("Log file:", logFile.Name())
		cmd.Stderr = logFile
		cmd.Stdout = logFile
	}

	return cmd.Run()
}

func Download(source, file string) error {
	log.Println("Downloading", source)

	out, err := os.Create(file)
	if err != nil {
		return err
	}

	resp, err := http.Get(source)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	resp.Body.Close()

	log.Println("Downloaded", file)

	return nil
}

func GetURLBody(url string) (string, error) {
	log.Println("Retrieving URL Body of", url)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	resp.Body.Close()

	return string(body), nil
}

func UnzipFolder(source string, destDir string) error {
	log.Println("Extracting", source)

	zip, err := zip.OpenReader(source)
	if err != nil {
		return err
	}

	for _, file := range zip.File {
		filePath := filepath.Join(destDir, file.Name)
		log.Println("Unzipping", filePath)

		if file.FileInfo().IsDir() {
			CreateDirs(filePath)

			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
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
