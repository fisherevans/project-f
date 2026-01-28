//go:build windows

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var logFile *os.File

func redirectLogsToFile(path string) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create log directory %q: %v\n", dir, err)
		return
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file %q for writing: %v\n", path, err)
		return
	}

	logFile = f
	log.SetOutput(logFile)
}
