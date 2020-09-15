package spruce

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/starkandwayne/goutils/tree"

	. "github.com/geofffranks/spruce/log"
)

// ShuffleOperator ...
type ShuffleOperator struct{}

// Setup ...
func (ShuffleOperator) Setup() error {
	return nil
}

// Phase ...
func (ShuffleOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (ShuffleOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor, auto []*tree.Cursor) []*tree.Cursor {
	return auto
}

// Run ...
func (ShuffleOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( shuffle ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( shuffle ... )) operation at $%s\n", ev.Here)

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

			switch s.(type) {
			case []interface{}:
				for _, thing := range s.([]interface{}) {
					vals = append(vals, thing)
				}

			case map[interface{}]interface{}:
				DEBUG("     [%d]: resolved to a map; error!", i)
				return nil, fmt.Errorf("shuffle only accepts arrays and string values")

			default:
				vals = append(vals, s.(interface{}))
			}

		default:
			DEBUG("  arg[%d]: I don't know what to do with '%v'", i, arg)
			return nil, fmt.Errorf("shuffle operator only accepts key reference arguments")
		}
		DEBUG("")
	}

	return &Response{
		Type:  Replace,
		Value: shuffle(vals),
	}, nil
}

func init() {
	RegisterOp("shuffle", ShuffleOperator{})
}

func shuffle(l []interface{}) []interface{} {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(l), func(i, j int) { l[i], l[j] = l[j], l[i] })
	return l
}
