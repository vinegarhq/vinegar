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
	// prefix-2006-01-02T15:04:05Z07:00.log
	file, err := os.Create(filepath.Join(Dirs.Log, prefix+"-"+time.Now().Format(time.RFC3339)+".log"))
	if err != nil {
		log.Fatalf("failed to create %s log file: %w", prefix, err)
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

	if err := os.RemoveAll(source); err != nil {
		return err
	}

	return nil
}

// Loop over all proc(5)fs PID directories and check if the given query (string)
// matches the file contents of with a file called 'comm', within the PID
// directory. For simplification purposes this will use a /proc/*/comm glob instead.
// Once found a 'comm' file, simply return true; return false when not found.
func CommFound(query string) bool {
	comms, err := filepath.Glob("/proc/*/comm")
	if err != nil {
		log.Fatal("failed to locate procfs commands")
	}

	for _, comm := range comms {
		c, err := os.ReadFile(comm)
		// The 'comm' file contains a new line, we remove it, as it will mess up
		// the query. hence 'minus'ing the length by 1 removes the newline.
		if err == nil && string(c)[:len(c)-1] == query {
			return true
		}
	}

	return false
}

// Simply loop for every second to see if a process query 'comm' has not been
// found, or in other words has exited. this function will simply stop the current
// execution queue and simply just waits, and the functions following this one will
// execute.
func CommLoop(comm string) {
	log.Println("Waiting for process command:", comm)

	for {
		time.Sleep(time.Second)

		if !CommFound(comm) {
			break
		}
	}
}
