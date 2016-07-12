package spruce

import (
	"fmt"
	"github.com/starkandwayne/goutils/ansi"
	"sort"
	"strings"
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
