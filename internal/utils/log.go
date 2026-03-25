package utils

import (
	"fmt"
	"time"
)

// Log outputs a timestamped log message
// If multiple arguments are provided, the first is treated as a format string
func Log(format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}
	fmt.Printf("%s %s\n", timestamp, msg)
}
