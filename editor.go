package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

func GetEditor() (string, error) {
	editor, ok := os.LookupEnv("EDITOR")

	if !ok {
		return "", errors.New("no $EDITOR variable set")
	}

	if _, err := exec.LookPath(editor); err != nil {
		return "", fmt.Errorf("invalid $EDITOR: %w", err)
	}

	return editor, nil
}

func EditConfig() {
	var testConfig Configuration

	editor, err := GetEditor()
	if err != nil {
		log.Fatal("unable to find editor: ", err)
	}

	tempConfigFile, err := os.CreateTemp(Dirs.Config, "testconfig.*.toml")
	if err != nil {
		log.Fatal(err)
	}

	tempConfigFilePath, err := filepath.Abs(tempConfigFile.Name())
	if err != nil {
		log.Fatal(err)
	}

	configFile, err := os.ReadFile(ConfigFilePath)
	if err != nil {
		log.Fatal(err)
	}

	if _, err = tempConfigFile.Write(configFile); err != nil {
		log.Fatal(err)
	}

	tempConfigFile.Close()

	editorCmd := exec.Command(editor, tempConfigFilePath)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stderr = os.Stderr
	editorCmd.Stdout = os.Stdout

	for {
		if err := editorCmd.Run(); err != nil {
			log.Fatal(err)
		}

		if _, err := toml.DecodeFile(tempConfigFilePath, &testConfig); err != nil {
			log.Println(err)
			log.Println("Press enter to re-edit configuration file")
			fmt.Scanln()

			continue
		}

		if err := os.Rename(tempConfigFilePath, ConfigFilePath); err != nil {
			log.Fatal(err)
		}

		break
	}
}
