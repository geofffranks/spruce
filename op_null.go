package main

import (
	"fmt"
)

// NullOperator ...
type NullOperator struct {
	Missing string
}

// Setup ...
func (NullOperator) Setup() error {
	return nil
}

// Phase ...
func (NullOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (NullOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*Cursor) []*Cursor {
	return []*Cursor{}
}

// Run ...
func (n NullOperator) Run(ev *Evaluator, _ []*Expr) (*Response, error) {
	return nil, fmt.Errorf("(( %s )) operator not defined", n.Missing)
}
