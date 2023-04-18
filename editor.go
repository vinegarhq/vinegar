package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/BurntSushi/toml"
)

// 'Easy and convenient' way to edit Vinegar's configuration, which uses
// a temporary configuration file for changes, and uses the $EDITOR variable.
func EditConfig() {
	var testConfig Configuration

	temp, err := TempConfigFile()
	if err != nil {
		log.Fatalf("failed to create temp config file: %s", err)
	}

	for {
		if err := ExecEditor(temp); err != nil {
			// Immediately remove the temp file after failure
			// of editing the file.
			os.Remove(temp)

			log.Fatal(err)
		}

		if _, err := toml.DecodeFile(temp, &testConfig); err != nil {
			log.Println(err)
			log.Println("Press enter to re-edit configuration file")
			fmt.Scanln()

			continue
		}

		// After the temporary configuration file has no errors,
		// move it to the global configuration file.
		if err := os.Rename(temp, ConfigFilePath); err != nil {
			log.Fatal(err)
		}

		break
	}
}

// Launches $EDITOR with the given filepath. $EDITOR is used here
// as a way to retrieve the system's editor, as it is a great standard
// used by many applications (also POSIX specifies it).
func ExecEditor(filePath string) error {
	editor, ok := os.LookupEnv("EDITOR")

	if !ok {
		return errors.New("no $EDITOR variable set")
	}

	cmd := exec.Command(editor, filePath)

	// Required, causes problems with the editor otherwise.
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func TempConfigFile() (string, error) {
	temp, err := os.CreateTemp(Dirs.Config, "testconfig.*.toml")
	if err != nil {
		return "", err
	}
	defer temp.Close()

	// Open the configuration file, for copying it's contents
	// to the temporary configuration file.
	config, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		return "", err
	}

	// We use CreateTemp, and this is one of (if not) the only
	// ways to copy the configuration file's contents.
	if _, err = temp.Write(config); err != nil {
		return "", err
	}

	return temp.Name(), nil
}
