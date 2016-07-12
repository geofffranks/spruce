package spruce

import (
	"fmt"
	"github.com/starkandwayne/goutils/tree"
)

// ParamOperator ...
type ParamOperator struct{}

// Setup ...
func (ParamOperator) Setup() error {
	return nil
}

// Phase ...
func (ParamOperator) Phase() OperatorPhase {
	return ParamPhase
}

// Dependencies ...
func (ParamOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor) []*tree.Cursor {
	return []*tree.Cursor{}
}

// Run ...
func (ParamOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	v, _ := args[0].Evaluate(ev.Tree) // FIXME: there are lots of assumptions here...
	return nil, fmt.Errorf("%s", v)
}

func init() {
	RegisterOp("param", ParamOperator{})
}
