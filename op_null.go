package main

import (
	"fmt"
)

type NullOperator struct{
	Missing string
}

func (NullOperator) Dependencies(_ *Evaluator, _ []interface{}, _ []*Cursor) []*Cursor {
	return []*Cursor{}
}

func (n NullOperator) Run(ev *Evaluator, args []interface{}) (*Response, error) {
	return nil, fmt.Errorf("(( %s )) operator not defined", n.Missing)
}
