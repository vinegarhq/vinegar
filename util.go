// Copyright vinegar-development 2023

package main

import (
	"archive/zip"
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var InFlatpak bool = InFlatpakCheck()

// Helper function to handle error failure
func Errc(e error, message ...string) {
	if e != nil {
		if message != nil {
			log.Println(message)
		}
		panic(e)
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
func CheckDirs(perm uint32, dir ...string) {
	for _, d := range dir {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			log.Println("Creating directory", d, "with permissions", os.FileMode(perm))
		} else {
			continue
		}
		Errc(os.MkdirAll(d, os.FileMode(perm)))
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

	cmd.Stdin  = os.Stdin
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

// Unzip a single file without keeping track of zip's structure into
// a target file, Will remove the source zip file after successful 
// extraction, this function only works for zips with a single file,
// otherwise it will overwrie the target with other files in the zip.
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
// NOTE: this WILL not work on FreeBSD.
func CommFound(query string) bool {
	comms, err := filepath.Glob("/proc/*/comm")
	Errc(err)

	for _, comm := range comms {
		// comm file will include newline by default, we just remove it
		c, err := os.ReadFile(comm)
		if err == nil && string(c)[:len(c)-1] == query {
			return true
		}
	}

	return false
}

// Stop the current execution queue and wait until a pid with the
// comm has exited or has been killed.
func CommLoop(comm string) {
	log.Println("Waiting for process named", comm, "to exit")

	// wait a bit for the process to start
	time.Sleep(time.Second)

	for {
		time.Sleep(time.Second)

		if !CommFound(comm) {
			break
		}
	}
}

// Launch the system's editor $EDITOR, if not found, use xdg-open.
// Aditionally, apply the changes to a temporary configuration file
// and write it when the temporary is successful and can be parsed.
func EditConfig() {
	var editor string
	var testConfig Configuration

	editorVar := os.Getenv("EDITOR")

	if editorVar != "" {
		editor = editorVar
	} else if _, e := exec.LookPath("xdg-open"); e == nil {
		editor = "xdg-open"
	} else {
		log.Fatal("Failed to find editor")
	}

	// Create a temporary configuration file for testing
	tempConfigFile, err := os.CreateTemp(Dirs.Config, "testconfig.*.toml")
	Errc(err)

	// Absolute path is required for removal and editing
	tempConfigFilePath, err := filepath.Abs(tempConfigFile.Name())
	Errc(err)

	configFile, err := os.ReadFile(ConfigFilePath)
	Errc(err)

	// Write the original configuration file to the temporary one
	_, err = tempConfigFile.Write(configFile)
	Errc(err)

	// Loop until toml.Unmarshal is successful
	for {
		Exec(editor, false, tempConfigFilePath)

		tempConfig, err := os.ReadFile(tempConfigFilePath)
		Errc(err)

		// Check if parsing has _failed_, if failed; re-open the file, which instructs
		// contiuation of the loop.
		if err := toml.Unmarshal([]byte(tempConfig), &testConfig); err != nil {
			log.Println(err.Error())
			log.Println("Press enter to re-edit configuration file")
			fmt.Scanln()
			continue
		}

		// If parsing has been successful ^, move the
		// temporary configuration file to the primary one.
		Errc(os.Rename(tempConfigFilePath, ConfigFilePath))
		break
	}
}
