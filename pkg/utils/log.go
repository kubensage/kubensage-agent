package utils

import (
	"log"
	"os"
)

// SetupLogging configures the global logger to write output to the specified file.
//
// The caller is responsible for closing the returned file (typically via `defer`).
// If the file cannot be opened, the function logs a fatal error and exits.
func SetupLogging(logFileName string) *os.File {
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}

	log.SetOutput(logFile)
	return logFile
}
