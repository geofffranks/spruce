package utils

import (
	"fmt"
	"os"
	"strings"
)

var Debug = false

// DEBUG - Prints out a debug message
func DEBUG(format string, args ...interface{}) {
	if Debug {
		content := fmt.Sprintf(format, args...)
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			lines[i] = "DEBUG> " + line
		}
		content = strings.Join(lines, "\n")
		fmt.Fprintf(os.Stderr, format, args...)
	}
}
