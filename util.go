package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	log.Println("Downloading", source, "to", file)

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

// Unzip a single target file in the source zip file to a file,
// and keep it's permissions, afterwards; remove the source zip file.
// this is ONLY used for extracting rbxfpsunlocker,
// as it will ignore all other files.
func UnzipFile(source, target, file string) error {
	log.Println("Unzipping:", source)

	zip, err := zip.OpenReader(source)
	if err != nil {
		return err
	}

	for _, zipped := range zip.File {
		if zipped.Name != target {
			continue
		}

		reader, err := zipped.Open()
		if err != nil {
			return err
		}

		target, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipped.Mode())
		if err != nil {
			return err
		}

		if _, err := io.Copy(target, reader); err != nil {
			return err
		}

		log.Println("Unzipped:", zipped.Name)
	}

	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("target unzip file does not exist")
	}

	log.Println("Removing source zip file")

	return os.RemoveAll(source)
}
