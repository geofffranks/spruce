package utils

import (
	"fmt"
	"os"
)

func PrintfStdOut(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format, args...)
}

func PrintfStdErr(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}
