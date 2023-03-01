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
	"time"
)

func LogFile(prefix string) *os.File {
	file, err := os.Create(filepath.Join(Dirs.Log, prefix+"-"+time.Now().Format(time.RFC3339)+".log"))
	if err != nil {
		log.Fatal(err)
	}

	return file
}

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

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func Download(source, file string) error {
	log.Println("Downloading:", source)

	out, err := os.Create(file)
	if err != nil {
		return err
	}

	resp, err := http.Get(source)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

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

	if err := os.RemoveAll(source); err != nil {
		return err
	}

	return nil
}

func CommFound(query string) bool {
	comms, err := filepath.Glob("/proc/*/comm")
	if err != nil {
		log.Fatal("failed to locate procfs commands")
	}

	for _, comm := range comms {
		c, err := os.ReadFile(comm)
		if err == nil && string(c)[:len(c)-1] == query {
			return true
		}
	}

	return false
}

func CommLoop(comm string) {
	log.Println("Waiting for process command:", comm)

	for {
		time.Sleep(time.Second)

		if !CommFound(comm) {
			break
		}
	}
}
