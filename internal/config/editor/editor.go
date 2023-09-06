package editor

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/BurntSushi/toml"
	"github.com/vinegarhq/vinegar/internal/config"
	"github.com/vinegarhq/vinegar/internal/dirs"
)

func EditConfig() {
	var cfg config.Config

	editor, err := Editor()
	if err != nil {
		log.Fatalf("failed to find editor: %s", err)
	}

	cmd := exec.Command(editor, config.Path)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := dirs.Mkdirs(dirs.Config); err != nil {
		log.Fatal(err)
	}

	file, err := os.OpenFile(config.Path, os.O_WRONLY|os.O_CREATE, 0o644)
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

	if info.Size() < 1 {
		log.Println("Writing Configuration template")

		if _, err := file.WriteString(template); err != nil {
			log.Fatal(err)
		}
	}

	file.Close()

	for {
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}

		if _, err := toml.DecodeFile(config.Path, &cfg); err != nil {
			log.Println(err)
			log.Println("Press enter to re-edit configuration file")
			fmt.Scanln()

			continue
		}

		break
	}
}

func Editor() (string, error) {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor, nil
	}

	log.Println("no EDITOR set, falling back to nano")

	return exec.LookPath("nano")
}
