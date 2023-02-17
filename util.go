// Copyright vinegar-development 2023

package main

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var InFlatpak bool = InFlatpakCheck()

// Helper function to handle error failure
func Errc(e error, message ...string) {
	if e != nil {
		if message != nil {
			log.Println(message)
		}
		log.Fatal(e)
	}
}

// Deletes directories, but with logging(!)
func DeleteDirs(dir ...string) {
	for _, d := range dir {
		log.Println("Deleting directory:", d)
		Errc(os.RemoveAll(d))
	}
}

// Check for directories if they exist, if not,
// create them with 0755, and let the user know with logging.
func CheckDirs(dir ...string) {
	for _, d := range dir {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			log.Println("Creating directory:", d)
		} else {
			continue
		}
		Errc(os.MkdirAll(d, 0755))
	}
}

// Execute a program with arguments whilst keeping
// it's stderr output to a log file, stdout is ignored and is sent to os.Stdout.
func Exec(prog string, logStderr bool, args ...string) {
	log.Println("Executing:", prog, args) // debug

	cmd := exec.Command(prog, args...)
	cmd.Dir = Dirs.Cache

	// Stdout is particularly always empty.
	if logStderr {
		stderrFile, err := os.Create(filepath.Join(Dirs.Log, time.Now().Format(time.RFC3339)+"-stderr.log"))
		Errc(err)
		log.Println("Forwarding stderr to", stderrFile.Name())
		defer stderrFile.Close()
		cmd.Stderr = stderrFile
	} else {
		cmd.Stderr = os.Stderr
	}

	cmd.Stdin = os.Stdin //Fix for nano bug in "vinegar edit"
	cmd.Stdout = os.Stdout

	Errc(cmd.Run())
}

// Download a single file into a target file.
func Download(source, target string) {
	// Create blank file
	out, err := os.Create(target)
	Errc(err)
	defer out.Close()

	resp, err := http.Get(source)
	Errc(err)
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	Errc(err)

	log.Println("Downloaded", source)
}

// Unzip a single file without keeping track of zip's structue into
// a target file, Will remove the source zip file after successful extraction.
func Unzip(source, target string) {
	archive, err := zip.OpenReader(source)
	Errc(err)
	defer archive.Close()

	// Create blank file
	out, err := os.Create(target)
	Errc(err)
	defer out.Close()

	for _, zipped := range archive.File {
		zippedFile, err := zipped.Open()
		Errc(err)
		defer zippedFile.Close()

		_, err = io.Copy(out, zippedFile)
		Errc(err)

		log.Println("Unzipped", zipped.Name)
	}

	Errc(os.RemoveAll(source))
	log.Println("Removed archive", source)
}

// Check if running in flatpak (in which case it is necessary to disable DXVK)
func InFlatpakCheck() bool {
	if _, err := os.Stat("/.flatpak-info"); err != nil {
		return false
	} else {
		return true
	}
}

// Loop over procfs (/proc) for if pid/comm matches a string, when found
// such process, return true; false otherwise
func CommFound(comm string) bool {
	procDir, err := os.Open("/proc")
	Errc(err)

	procs, err := procDir.Readdir(0)
	Errc(err)

	for _, p := range procs {
		procComm, _ := os.ReadFile(filepath.Join(procDir.Name(), p.Name(), "comm"))
		if strings.HasPrefix(string(procComm), comm) {
			return true
		}
	}

	return false
}

// Loop until comm has not been found, then execute provided function
func CommLoop(comm string, action func()) {
	log.Println("Waiting for process named", comm, "to exit")

	for {
		time.Sleep(time.Second)

		if !CommFound(comm) {
			break
		}
	}

	action()
}

// Launch the system's editor $EDITOR, if not found, use xdg-open
func EditConfig() {
	editorVar := os.Getenv("EDITOR")

	if editorVar != "" {
		Exec(editorVar, false, ConfigFilePath)
	} else if _, e := exec.LookPath("xdg-open"); e == nil {
		Exec("xdg-open", false, ConfigFilePath)
	} else {
		log.Fatal("Failed to find editor")
	}
}
