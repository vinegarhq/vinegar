package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Exec(prog string, logName string, args ...string) error {
	if prog == "wine" {
		PfxInit()
	}

	log.Println("Executing:", prog, args)

	cmd := exec.Command(prog, args...)

	cmd.Dir = Dirs.Cwd
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if logName != "" {
		logFile := LogFile(logName)
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
		// Roblox's Zip Files have windows paths in them
		filePath := filepath.Join(destDir, strings.ReplaceAll(file.Name, `\`, "/"))
		log.Println("Unzipping", filePath)

		if file.FileInfo().IsDir() {
			if err := os.Mkdir(filePath, file.Mode()); err != nil {
				return err
			}

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

func UntarGzipFolder(source string, destDir string) error {
	log.Println("Extracting", source)

	tarball, err := os.Open(source)
	if err != nil {
		return err
	}

	stream, err := gzip.NewReader(tarball)
	if err != nil {
		return err
	}

	tar := tar.NewReader(stream)

	for {
		header, err := tar.Next()

		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		filePath := filepath.Join(destDir, header.Name)
		info := header.FileInfo()

		log.Println("Ungzipping", filePath)

		if info.IsDir() {
			if err := os.Mkdir(filePath, info.Mode()); err != nil {
				return err
			}

			continue
		}

		destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}

		if _, err = io.Copy(destFile, tar); err != nil {
			return err
		}

		destFile.Close()
	}

	return nil
}

func VerifyFileMD5(filePath string, signature string) {
	log.Printf("Verifying file %s: %s", filePath, signature)

	hash := md5.New()

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal(err)
	}

	if signature != hex.EncodeToString(hash.Sum(nil)) {
		log.Fatalf("File %s checksum mismatch: %x", filePath, hash.Sum(nil))
	}

	file.Close()
}
