package log

import (
	"fmt"
	"os"
	"strings"
)

var DebugOn bool = false
var TraceOn bool = false

var printfStdErr func(string, ...interface{})

func init() {
	printfStdErr = func(format string, args ...interface{}) {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

func DEBUG(format string, args ...interface{}) {
	if DebugOn {
		content := fmt.Sprintf(format, args...)
		lines := strings.Split(content, "\n")
		for i, line := range lines {
			lines[i] = "DEBUG> " + line
		}
		content = strings.Join(lines, "\n")
		printfStdErr("%s\n", content)
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
		printfStdErr("%s\n", content)
	}
}
