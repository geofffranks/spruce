package spruce

import (
	"fmt"

	"github.com/geofffranks/spruce/log"
	"github.com/starkandwayne/goutils/tree"
)

// NegateOperator ...
type NegateOperator struct{}

// Setup ...
func (NegateOperator) Setup() error {
	return nil
}

// Phase ...
func (NegateOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (NegateOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor, auto []*tree.Cursor) []*tree.Cursor {
	return auto
}

// Run ...
func (NegateOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	log.DEBUG("running (( negate ... )) operation at $.%s", ev.Here)
	defer log.DEBUG("done with (( negate ... )) operation at $%s\n", ev.Here)

	if len(args) != 1 {
		return nil, fmt.Errorf("negate operator requires exactly one reference argument")
	}

	var arg = args[0]
	var val bool
	v, err := arg.Resolve(ev.Tree)
	if err != nil {
		log.DEBUG(" resolution failed\n error: %s", err)
		return nil, err
	}
	switch v.Type {
	case Reference:
		log.DEBUG(" trying to resolve reference $.%s", v.Reference)
		s, err := v.Reference.Resolve(ev.Tree)
		if err != nil {
			log.DEBUG(" resolution failed\n error: %s", err)
			return nil, fmt.Errorf("Unable to resolve `%s`: %s", v.Reference, err)
		}
		log.DEBUG("  resolved to a value")
		switch s2 := s.(type) {
		case bool:
			val = !s2
		default:
			return nil, fmt.Errorf("negate operator only accepts references to bools")
		}
	case Literal:
		switch literal := v.Literal.(type) {
		case bool:
			val = !literal
		default:
			return nil, fmt.Errorf("negate operator only operates on bools")
		}

	default:
		log.DEBUG(" unsupported expression type %v, only references are allowed: '%v'", v.Type, arg)
		return nil, fmt.Errorf("negate operator only accepts reference arguments")
	}

	return &Response{
		Type:  Replace,
		Value: val,
	}, nil
}

func init() {
	RegisterOp("negate", NegateOperator{})
}
