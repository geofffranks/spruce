package main

import (
	"fmt"
	"strings"
	"sort"
)

type MultiError struct {
	Errors []error
}

func (e MultiError) Error() string {
	s := []string{}
	for _, err := range e.Errors {
		s = append(s, fmt.Sprintf(" - %s\n", err))
	}

	sort.Strings(s)
	return fmt.Sprintf("%d error(s) detected:\n%s\n", len(e.Errors), strings.Join(s, ""))
}

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

type SyntaxError struct {
	Problem  string
	Position int
}

func (e SyntaxError) Error() string {
	return fmt.Sprintf("syntax error: %s at position %d", e.Problem, e.Position)
}

type TypeMismatchError struct {
	Path   []string
	Wanted string
	Got    string
	Value  interface{}
}

func (e TypeMismatchError) Error() string {
	if e.Got == "" {
		return fmt.Sprintf("%s is not", strings.Join(e.Path, "."), e.Wanted)
	} else {
		if e.Value != nil {
			return fmt.Sprintf("$.%s [=%v] is %s (not %s)", strings.Join(e.Path, "."), e.Value, e.Got, e.Wanted)
		} else {
			return fmt.Sprintf("$.%s is %s (not %s)", strings.Join(e.Path, "."), e.Got, e.Wanted)
		}
	}
}

type NotFoundError struct {
	Path []string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("`$.%s` could not be found in the YAML datastructure", strings.Join(e.Path, "."))
}
