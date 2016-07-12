package spruce

import (
	"fmt"
	"sort"

	"github.com/starkandwayne/goutils/ansi"

	. "github.com/geofffranks/spruce/log"
	"github.com/starkandwayne/goutils/tree"
)

// KeysOperator ...
type KeysOperator struct{}

// Setup ...
func (KeysOperator) Setup() error {
	return nil
}

// Phase ...
func (KeysOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (KeysOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor) []*tree.Cursor {
	return []*tree.Cursor{}
}

// Run ...
func (KeysOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( keys ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( keys ... )) operation at $%s\n", ev.Here)

	var vals []string

	for i, arg := range args {
		v, err := arg.Resolve(ev.Tree)
		if err != nil {
			DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
			return nil, err
		}

		switch v.Type {
		case Literal:
			DEBUG("  arg[%d]: found string literal '%s'", i, v.Literal)
			DEBUG("           (keys operator only handles references to other parts of the YAML tree)")
			return nil, fmt.Errorf("keys operator only accepts key reference arguments")

		case Reference:
			DEBUG("  arg[%d]: trying to resolve reference $.%s", i, v.Reference)
			s, err := v.Reference.Resolve(ev.Tree)
			if err != nil {
				DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
				return nil, fmt.Errorf("Unable to resolve `%s`: %s", v.Reference, err)
			}

			m, ok := s.(map[interface{}]interface{})
			if !ok {
				DEBUG("     [%d]: resolved to something that is not a map.  that is unacceptable.", i)
				return nil, ansi.Errorf("@c{%s} @R{is not a map}", v.Reference)
			}
			DEBUG("     [%d]: resolved to a map; extracting keys", i)
			for k := range m {
				vals = append(vals, k.(string))
			}

		default:
			DEBUG("  arg[%d]: I don't know what to do with '%v'", i, arg)
			return nil, fmt.Errorf("keys operator only accepts key reference arguments")
		}
		DEBUG("")
	}

	switch len(args) {
	case 0:
		DEBUG("  no arguments supplied to (( keys ... )) operation.  oops.")
		return nil, ansi.Errorf("no arguments specified to @c{(( keys ... ))}")

	default:
		sort.Strings(vals)
		return &Response{
			Type:  Replace,
			Value: vals,
		}, nil
	}
}

func init() {
	RegisterOp("keys", KeysOperator{})
}
