package tree

import (
	"fmt"
	"github.com/starkandwayne/goutils/ansi"
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
		return ansi.Sprintf("@c{%s} @R{is not} @m{%s}", strings.Join(e.Path, "."), e.Wanted)
	}
	if e.Value != nil {
		return ansi.Sprintf("@c{$.%s} @R{[=%v] is %s (not} @m{%s}@R{)}", strings.Join(e.Path, "."), e.Value, e.Got, e.Wanted)
	}
	return ansi.Sprintf("@C{$.%s} @R{is %s (not} @m{%s}@R{)}", strings.Join(e.Path, "."), e.Got, e.Wanted)
}

// NotFoundError ...
type NotFoundError struct {
	Path []string
}

// Error ...
func (e NotFoundError) Error() string {
	return ansi.Sprintf("@R{`}@c{$.%s}@R{` could not be found in the datastructure}", strings.Join(e.Path, "."))
}
