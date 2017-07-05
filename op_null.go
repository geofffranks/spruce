package spruce

import (
	"github.com/starkandwayne/goutils/ansi"
	"github.com/starkandwayne/goutils/tree"
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
func (NullOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor, _ []*tree.Cursor) []*tree.Cursor {
	return nil
}

// Run ...
func (n NullOperator) Run(ev *Evaluator, _ []*Expr) (*Response, error) {
	return nil, ansi.Errorf("@c{(( %s ))} @R{operator not defined}", n.Missing)
}
