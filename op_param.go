package spruce

import (
	"fmt"

	"github.com/starkandwayne/goutils/ansi"
	"github.com/starkandwayne/goutils/tree"

	. "github.com/geofffranks/spruce/log"
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
func (ParamOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor, _ []*tree.Cursor) []*tree.Cursor {
	return nil
}

// Run ...
func (ParamOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( param ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( param ... )) operation at $%s\n", ev.Here)

	switch len(args) {
	case 2:
		def, err := args[1].Evaluate(ev.Tree)
		if err != nil {
			return nil, err
		}
		return &Response{
			Type:  Replace,
			Value: def,
		}, nil
	case 1:
		v, _ := args[0].Evaluate(ev.Tree) // FIXME: there are lots of assumptions here...
		return nil, fmt.Errorf("%s", v)
	default:
		return nil, ansi.Errorf("wrong number of arguments specified to @c{(( param ... ))}")
	}
}

func init() {
	RegisterOp("param", ParamOperator{})
}
