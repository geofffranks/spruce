package tree

import (
	"fmt"
	"strings"
)

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
	return fmt.Sprintf("`$.%s` could not be found in the datastructure", strings.Join(e.Path, "."))
}
