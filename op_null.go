package main

import (
	"fmt"
)

// NullOperator ...
type NullOperator struct {
	Missing string
}

// Dependencies ...
func (NullOperator) Dependencies(_ *Evaluator, _ []interface{}, _ []*Cursor) []*Cursor {
	return []*Cursor{}
}

// Run ...
func (n NullOperator) Run(ev *Evaluator, args []interface{}) (*Response, error) {
	return nil, fmt.Errorf("(( %s )) operator not defined", n.Missing)
}
