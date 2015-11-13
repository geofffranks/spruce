package main

import (
	"fmt"
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
	return fmt.Sprintf("%d error(s) detected:\n%s\n", len(e.Errors), strings.Join(s, ""))
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

// SyntaxError ...
type SyntaxError struct {
	Problem  string
	Position int
}

// Error ...
func (e SyntaxError) Error() string {
	return fmt.Sprintf("syntax error: %s at position %d", e.Problem, e.Position)
}

// TypeMismatchError ...
type TypeMismatchError struct {
	Path   []string
	Wanted string
	Got    string
	Value  interface{}
}

// Error ...
func (e TypeMismatchError) Error() string {
	if e.Got == "" {
		return fmt.Sprintf("%s is not %s", strings.Join(e.Path, "."), e.Wanted)
	}
	if e.Value != nil {
		return fmt.Sprintf("$.%s [=%v] is %s (not %s)", strings.Join(e.Path, "."), e.Value, e.Got, e.Wanted)
	}
	return fmt.Sprintf("$.%s is %s (not %s)", strings.Join(e.Path, "."), e.Got, e.Wanted)
}

// NotFoundError ...
type NotFoundError struct {
	Path []string
}

// Error ...
func (e NotFoundError) Error() string {
	return fmt.Sprintf("`$.%s` could not be found in the YAML datastructure", strings.Join(e.Path, "."))
}
