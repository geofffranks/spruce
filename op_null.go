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
func (NullOperator) Dependencies(_ *Evaluator, _ []interface{}, _ []*Cursor) []*Cursor {
	return []*Cursor{}
}

// Run ...
func (n NullOperator) Run(ev *Evaluator, args []interface{}) (*Response, error) {
	return nil, fmt.Errorf("(( %s )) operator not defined", n.Missing)
}
