package spruce

import (
	"encoding/base64"
	"fmt"

	"github.com/starkandwayne/goutils/ansi"

	"github.com/starkandwayne/goutils/tree"

	. "github.com/geofffranks/spruce/log"
)

// Base64DecodeOperator ...
type Base64DecodeOperator struct{}

// Setup ...
func (Base64DecodeOperator) Setup() error {
	return nil
}

// Phase ...
func (Base64DecodeOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (Base64DecodeOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor, auto []*tree.Cursor) []*tree.Cursor {
	return auto
}

// Run ...
func (Base64DecodeOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( base64-decode ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( base64-decode ... )) operation at $%s\n", ev.Here)

	if len(args) != 1 {
		return nil, fmt.Errorf("base64-decode operator requires exactly one string or reference argument")
	}

	var contents string

	arg := args[0]
	i := 0
	v, err := arg.Resolve(ev.Tree)
	if err != nil {
		DEBUG("  arg[%d]: failed to resolve expression to a concrete value", i)
		DEBUG("     [%d]: error was: %s", i, err)
		return nil, err
	}

	switch v.Type {
	case Literal:
		DEBUG("  arg[%d]: using string literal '%v'", i, v.Literal)
		DEBUG("     [%d]: appending '%v' to resultant string", i, v.Literal)
		if fmt.Sprintf("%T", v.Literal) != "string" {
			return nil, ansi.Errorf("@R{tried to base64 decode} @c{%v}@R{, which is not a string scalar}", v.Literal)
		}
		contents = fmt.Sprintf("%v", v.Literal)

	case Reference:
		DEBUG("  arg[%d]: trying to resolve reference $.%s", i, v.Reference)
		s, err := v.Reference.Resolve(ev.Tree)
		if err != nil {
			DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
			return nil, fmt.Errorf("unable to resolve `%s`: %s", v.Reference, err)
		}

		switch s.(type) {
		case string:
			DEBUG("     [%d]: appending '%s' to resultant string", i, s)
			contents = fmt.Sprintf("%v", s)

		default:
			DEBUG("  arg[%d]: %v is not a string scalar", i, s)
			return nil, ansi.Errorf("@R{tried to base64 decode} @c{%v}@R{, which is not a string scalar}", v.Reference)
		}

	default:
		DEBUG("  arg[%d]: I don't know what to do with '%v'", i, arg)
		return nil, fmt.Errorf("base64-decode operator only accepts string literals and key reference argument")
	}
	DEBUG("")

	if decoded, err := base64.StdEncoding.DecodeString(contents); err == nil {
		DEBUG("  resolved (( base64-decode ... )) operation to the string:\n    \"%s\"", string(decoded))
		return &Response{
			Type:  Replace,
			Value: string(decoded),
		}, nil
	} else {
		return nil, fmt.Errorf("unable to base64 decode string %s: %s", contents, err)
	}
}

func init() {
	RegisterOp("base64-decode", Base64DecodeOperator{})
}
