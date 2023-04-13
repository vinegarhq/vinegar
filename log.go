package main

import (
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
	// prefix-2006-01-02T15:04:05Z07:00.log
	file, err := os.Create(filepath.Join(Dirs.Log, prefix+"-"+time.Now().Format(time.RFC3339)+".log"))
	if err != nil {
		log.Fatalf("failed to create %s log file: %s", prefix, err)
	}

	return file
}

func LatestLogFile(glob string) string {
	// Since filepath.Glob sorts numerically, the 'newest' log files
	// will always be last (hence why retrieveing the last array element
	// is used), as they contain the date they were created at.
	// On-top of this, it also sorts alphabetically, so we only check for
	// log files that match the pattern.
	LogFiles, _ := filepath.Glob(filepath.Join(Dirs.Log, glob))

	if len(LogFiles) < 1 {
		return ""
	}

	return LogFiles[len(LogFiles)-1]
}
