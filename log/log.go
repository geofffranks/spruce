// Package log provides debug and trace logging for spruce.
package log

import (
	"fmt"
	"io"
	"os"
	"strings"
)

var DebugOn bool = false
var TraceOn bool = false

// Output is the writer used by DEBUG and TRACE. Defaults to os.Stderr.
var Output io.Writer = os.Stderr

// DEBUG - Prints out a debug message
func DEBUG(format string, args ...interface{}) {
	if DebugOn {
		content := fmt.Sprintf(format, args...)
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			lines[i] = "DEBUG> " + line
		}
		content = strings.Join(lines, "\n")
		fmt.Fprintf(Output, "%s\n", content)
	}
}

// TRACE - Prints out a trace message
func TRACE(format string, args ...interface{}) {
	if TraceOn {
		content := fmt.Sprintf(format, args...)
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			lines[i] = "-----> " + line
		}
		content = strings.Join(lines, "\n")
		fmt.Fprintf(Output, "%s\n", content)
	}
}
