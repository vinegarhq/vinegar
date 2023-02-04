// Copyright vinegar-development 2023

package vinegar

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Helper function to handle error failure
func Errc(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// Deletes directories, with logging(!)
func DeleteDirs(dir ...string) {
	for _, d := range dir {
		log.Println("Deleting directory:", d)
		Errc(os.RemoveAll(d))
	}
}

// Check for directories if they exist, if not, create them with 0755,
// and let the user know with logging.
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
func Exec(prog string, args ...string) {
	log.Println(args)
	cmd := exec.Command(prog, args...)
	timeFmt := time.Now().Format(time.RFC3339)

	stderrFile, err := os.Create(filepath.Join(Dirs.Log, timeFmt+"-stderr.log"))
	Errc(err)
	log.Println("Forwarding stderr to", stderrFile.Name())
	defer stderrFile.Close()

	cmd.Dir = Dirs.Cache

	// Stdout is particularly always empty, so we forward its to ours
	cmd.Stdout = os.Stdout
	cmd.Stderr = stderrFile

	Errc(cmd.Run())

	logFile, err := stderrFile.Stat()
	Errc(err)
	if logFile.Size() == 0 {
		log.Println("Empty stderr log file detected, deleting")
		Errc(os.RemoveAll(stderrFile.Name()))
	}
}

// Download a single file into a target file.
func Download(source string, target string) {
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
func Unzip(source string, target string) {
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
