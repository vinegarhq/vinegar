package logs

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/vinegarhq/aubun/internal/dirs"
)

func File(name string) *os.File {
	if err := dirs.Mkdirs(dirs.Logs); err != nil {
		log.Println(err)
		return nil
	}

	// name-2006-01-02T15:04:05Z07:00.log
	path := filepath.Join(dirs.Logs, name+"-"+time.Now().Format(time.RFC3339)+".log")

	file, err := os.Create(path)
	if err != nil {
		log.Printf("Failed to create %s log file: %s", name, err)
		return nil
	}

	log.Printf("Logging to file: %s", path)

	return file
}
