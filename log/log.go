package log

import (
	"fmt"
	"os"
	"strings"
)

var DebugOn bool = false
var TraceOn bool = false

//PrintfStdErr is a configurable hook to print to error output
var PrintfStdErr func(string, ...interface{})

func init() {
	PrintfStdErr = func(format string, args ...interface{}) {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

// DEBUG - Prints out a debug message
func DEBUG(format string, args ...interface{}) {
	if DebugOn {
		content := fmt.Sprintf(format, args...)
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			lines[i] = "DEBUG> " + line
		}
		content = strings.Join(lines, "\n")
		PrintfStdErr("%s\n", content)
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
		PrintfStdErr("%s\n", content)
	}
}
