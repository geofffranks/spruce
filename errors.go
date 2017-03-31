package spruce

import (
	"fmt"
	"sort"
	"strings"

	"github.com/geofffranks/spruce/log"
	"github.com/starkandwayne/goutils/ansi"
)

// MultiError ...
type MultiError struct {
	Errors []error
}

// Error ...
func (e MultiError) Error() string {
	s := []string{}
	for _, err := range e.Errors {
		s = append(s, fmt.Sprintf(" - %s\n", err))
	}

	sort.Strings(s)
	return ansi.Sprintf("@r{%d} error(s) detected:\n%s\n", len(e.Errors), strings.Join(s, ""))
}

// Count ...
func (e *MultiError) Count() int {
	return len(e.Errors)
}

// Append ...
func (e *MultiError) Append(err error) {
	if err == nil {
		return
	}

	if mult, ok := err.(MultiError); ok {
		e.Errors = append(e.Errors, mult.Errors...)
	} else {
		e.Errors = append(e.Errors, err)
	}
}

//WarningError should produce a warning message to stderr if the context set for
// the error fits the context the error was caught in.
type WarningError struct {
	warning string
	context ErrorContext
}

//An ErrorContext is a flag or set of flags representing the contexts that
// an error should have a special meaning in.
type ErrorContext uint

//Bitwise-or these together to represent several contexts
const (
	eContextAll          = 0
	eContextDefaultMerge = 1 << iota
)

var dontPrintWarning bool

//NewWarningError returns a new WarningError object that has the given warning
// message and context(s) assigned. Assigning no context should mean that all
// contexts are active. Ansi library enabled.
func NewWarningError(context ErrorContext, warning string, args ...interface{}) (err WarningError) {
	err.warning = ansi.Sprintf(warning, args...)
	err.context = context
	return
}

//SilenceWarnings when called with true will make it so that warnings will not
// print when Warn is called. Calling it with false will make warnings visible
// again. Warnings will print by default.
func SilenceWarnings(should bool) {
	dontPrintWarning = should
}

//Error will return the configured warning message as a string
func (e WarningError) Error() string {
	return e.warning
}

//HasContext returns true if the WarningError was configured with the given context (or all).
// False otherwise.
func (e WarningError) HasContext(context ErrorContext) bool {
	return e.context == 0 || (context&e.context > 0)
}

//Warn prints the configured warning to stderr.
func (e WarningError) Warn() {
	if !dontPrintWarning {
		log.PrintfStdErr(ansi.Sprintf("@Y{warning:} %s\n", e.warning))
	}
}
