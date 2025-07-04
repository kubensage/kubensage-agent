package utils

import (
	"log"
	"os"
)

func SetupLogging(logFileName string) {
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		return
	}
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			log.Fatalf("error closing log file: %v", err)
		}
	}(logFile)

	log.SetOutput(logFile)
}
