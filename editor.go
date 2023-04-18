package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/BurntSushi/toml"
)

// 'Easy and convenient' way to edit Vinegar's configuration, which
// directly edits the _global configuration file_ for changes, and
// decodes it, to check for errors and allows the user to re-edit the file.
func EditConfig() {
	var testConfig Configuration

	if err := os.MkdirAll(Dirs.Config, 0o755); err != nil {
		log.Fatal(err)
	}

	// Open or create the configuration file
	file, err := os.OpenFile(ConfigFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		log.Fatal(err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		log.Fatal(err)
	}

	template := "# See how to configure Vinegar on the documentation website:\n" +
		"# https://vinegarhq.github.io/Configuration\n\n"

	// Write the template to the file if it is empty.
	if info.Size() < 1 {
		if _, err := file.WriteString(template); err != nil {
			log.Fatal(err)
		}
	}

	// Close the file as soon as we are done with it, as the editor
	// will open it next.
	file.Close()

	for {
		if err := ExecEditor(ConfigFilePath); err != nil {
			log.Fatal(err)
		}

		if _, err := toml.DecodeFile(ConfigFilePath, &testConfig); err != nil {
			log.Println(err)
			log.Println("Press enter to re-edit configuration file")
			fmt.Scanln()

			continue
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
