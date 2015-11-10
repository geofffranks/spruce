package main

import (
	"fmt"
)

// ParamOperator ...
type ParamOperator struct{}

// Dependencies ...
func (ParamOperator) Dependencies(_ *Evaluator, _ []interface{}, _ []*Cursor) []*Cursor {
	return []*Cursor{}
}

// Run ...
func (ParamOperator) Run(ev *Evaluator, args []interface{}) (*Response, error) {
	return nil, fmt.Errorf("%s", args[0])
}

func init() {
	RegisterOp("param", ParamOperator{})
}
