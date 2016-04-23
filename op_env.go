package main

import (
	"github.com/jhunt/tree"
	"fmt"
	"os"
)

type EnvOperator struct{}

func (EnvOperator) Setup() error {
	return nil
}

func (EnvOperator) Phase() OperatorPhase {
	return EvalPhase
}

func (EnvOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor) []*tree.Cursor {
	return []*tree.Cursor{}
}

func (EnvOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( env ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( env ... )) operation at $.%s\n", ev.Here)

	// syntax: (( env "$HOME" || "/home/user" ))
	if len(args) == 0 {
		DEBUG("  no arguments supplied to (( env ... )) operation.  oops.")
		return nil, fmt.Errorf("no arguments specified to (( env ... ))")
	}

	for i, arg := range args {
		v, err := arg.Resolve(ev.Tree)
		if err != nil {
			DEBUG("  arg[%d]: failed to resolve expression to a concrete value", i)
			DEBUG("     [%d]: error was: %s", i, err)
			return nil, err
		}

		switch v.Type {
		case Literal:
			s := os.ExpandEnv(fmt.Sprintf("%v", v.Literal))
			DEBUG("  arg[%d]: found string literal '%s'", i, v.Literal)
			DEBUG("     [%d]: expanded to '%s'", i, v.Literal, s)
			if s == "" {
				DEBUG("     [%d]: empty expansion detected; skipping", i)
				DEBUG("")
				continue
			}
			DEBUG("")
			return &Response{
				Type:  Replace,
				Value: s,
			}, nil

		default:
			DEBUG("  arg[%d]: I don't know what to do with '%v'", i, arg)
			return nil, fmt.Errorf("env operator only accepts key reference arguments")
		}
	}

	DEBUG("no suitable (non-empty) expansions were found; returning empty string ''")
	return &Response{
		Type:  Replace,
		Value: "",
	}, nil
}

func init() {
	RegisterOp("env", EnvOperator{})
}
