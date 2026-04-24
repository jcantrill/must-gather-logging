package utils

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

var (
	logFile *os.File
	logMux  sync.Mutex
)

// SetLogFile opens a log file for writing
func SetLogFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	logFile = f
	return nil
}

// CloseLogFile closes the log file if open
func CloseLogFile() error {
	if logFile != nil {
		return logFile.Close()
	}
	return nil
}

// Log outputs a timestamped log message to stdout and optionally to a log file
// If multiple arguments are provided, the first is treated as a format string
// Thread-safe for concurrent use
func Log(format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}
	output := fmt.Sprintf("%s %s\n", timestamp, msg)

	logMux.Lock()
	defer logMux.Unlock()

	// Write to stdout
	fmt.Print(output)

	// Write to log file if set
	if logFile != nil {
		io.WriteString(logFile, output)
	}
}

// LogRaw outputs a message without timestamp to stdout and optionally to a log file
// Used for command output that already has its own formatting
// Thread-safe for concurrent use
func LogRaw(msg string) {
	output := msg + "\n"

	logMux.Lock()
	defer logMux.Unlock()

	// Write to stdout
	fmt.Print(output)

	// Write to log file if set
	if logFile != nil {
		io.WriteString(logFile, output)
	}
}
