package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func LogToFile() {
	if !Config.Log {
		return
	}

	logOutput := io.MultiWriter(os.Stderr, LogFile("vinegar"))
	log.SetOutput(logOutput)
}

func LogFile(prefix string) *os.File {
	if err := os.MkdirAll(Dirs.Logs, 0o755); err != nil {
		log.Fatal(err)
	}

	// prefix-2006-01-02T15:04:05Z07:00.log
	file, err := os.Create(filepath.Join(Dirs.Logs, prefix+"-"+time.Now().Format(time.RFC3339)+".log"))
	if err != nil {
		log.Fatalf("failed to create %s log file: %s", prefix, err)
	}

	return file
}

func ListLogFiles() {
	logFiles, _ := filepath.Glob(filepath.Join(Dirs.Logs, "*.log"))
	for _, file := range logFiles {
		fmt.Println(file)
	}
}
