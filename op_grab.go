package spruce

import (
	"fmt"

	"github.com/starkandwayne/goutils/ansi"
	"github.com/starkandwayne/goutils/tree"

	. "github.com/geofffranks/spruce/log"
)

// GrabOperator ...
type GrabOperator struct{}

// Setup ...
func (GrabOperator) Setup() error {
	return nil
}

// Phase ...
func (GrabOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (GrabOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor, auto []*tree.Cursor) []*tree.Cursor {
	return auto
}

// Run ...
func (GrabOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( grab ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( grab ... )) operation at $%s\n", ev.Here)

	var vals []interface{}

	for i, arg := range args {
		v, err := arg.Resolve(ev.Tree)
		if err != nil {
			DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
			return nil, err
		}

		switch v.Type {
		case Literal:
			DEBUG("  arg[%d]: found string literal '%s'", i, v.Literal)
			vals = append(vals, v.Literal)

		case Reference:
			DEBUG("  arg[%d]: trying to resolve reference $.%s", i, v.Reference)
			s, err := v.Reference.Resolve(ev.Tree)
			if err != nil {
				DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
				return nil, fmt.Errorf("Unable to resolve `%s`: %s", v.Reference, err)
			}
			DEBUG("     [%d]: resolved to a value (could be a map, a list or a scalar); appending", i)
			vals = append(vals, s)

		default:
			DEBUG("  arg[%d]: I don't know what to do with '%v'", i, arg)
			return nil, fmt.Errorf("grab operator only accepts key reference arguments")
		}
		DEBUG("")
	}

	switch len(args) {
	case 0:
		DEBUG("  no arguments supplied to (( grab ... )) operation.  oops.")
		return nil, ansi.Errorf("no arguments specified to @c{(( grab ... ))}")

	case 1:
		DEBUG("  called with only one argument; returning value as-is")
		return &Response{
			Type:  Replace,
			Value: vals[0],
		}, nil

	default:
		DEBUG("  called with more than one arguments; flattening top-level lists into a single list")
		flat := []interface{}{}
		for i, lst := range vals {
			switch lst.(type) {
			case []interface{}:
				DEBUG("    [%d]: $.%s is a list; flattening it out", i, args[i].Reference)
				flat = append(flat, lst.([]interface{})...)
			default:
				DEBUG("    [%d]: $.%s is not a list; appending it as-is", i, args[i].Reference)
				flat = append(flat, lst)
			}
		}
		DEBUG("")

		return &Response{
			Type:  Replace,
			Value: flat,
		}, nil
	}
}

func init() {
	RegisterOp("grab", GrabOperator{})
}
