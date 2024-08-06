package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

var logFile *os.File

// Returns logger with configuration
func GetLogger(withCaller bool) zerolog.Logger {
	if gin.IsDebugging() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}

	// Get the absolute path for the log file
	absPath, err := filepath.Abs("./schedulerV2.log")
	if err != nil {
		fmt.Printf("Failed to get absolute path: %v\n", err)
		os.Exit(1)
	}

	// Open a file for logging
	logFile, err = os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}

	// Create a multi-writer that writes to both console and file
	multi := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}, logFile)

	// Create a new logger
	newLogger := zerolog.New(multi).With().Timestamp()

	if withCaller {
		newLogger = newLogger.Caller()
	}

	return newLogger.Logger()
}
