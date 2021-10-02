package spruce

import (
	"github.com/geofffranks/spruce/log"
	"github.com/geofffranks/yaml"
	fmt "github.com/starkandwayne/goutils/ansi"
	"github.com/starkandwayne/goutils/tree"
)

// StringifyOperator ...
type StringifyOperator struct{}

// Setup ...
func (StringifyOperator) Setup() error {
	return nil
}

// Phase ...
func (StringifyOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (StringifyOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor, auto []*tree.Cursor) []*tree.Cursor {
	return auto
}

// Run ...
func (StringifyOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	log.DEBUG("running (( stringify ... )) operation at $.%s", ev.Here)
	defer log.DEBUG("done with (( stringify ... )) operation at $%s\n", ev.Here)

	if len(args) != 1 {
		return nil, fmt.Errorf("stringify operator requires exactly one reference argument")
	}

	var arg = args[0]
	var val interface{}
	v, err := arg.Resolve(ev.Tree)
	if err != nil {
		log.DEBUG(" resolution failed\n error: %s", err)
		return nil, err
	}

	switch v.Type {
	case Literal:
		log.DEBUG(" found literal '%s'", v.Literal)
		val = v.Literal

	case Reference:
		log.DEBUG(" trying to resolve reference $.%s", v.Reference)
		s, err := v.Reference.Resolve(ev.Tree)
		if err != nil {
			log.DEBUG(" resolution failed\n error: %s", err)
			return nil, fmt.Errorf("Unable to resolve `%s`: %s", v.Reference, err)
		}
		log.DEBUG("  resolved to a value (could be a map, a list or a scalar)")
		data, err := yaml.Marshal(s)
		if err != nil {
			log.DEBUG("   marshaling failed\n   error: %s", err)
			return nil, fmt.Errorf("Unable to marshal `%s`: %s", v.Reference, err)
		}
		val = string(data)

	default:
		log.DEBUG(" unsupported expression type, only references are allowed: '%v'", arg)
		return nil, fmt.Errorf("stringify operator only accepts reference arguments")
	}
	log.DEBUG("")

	return &Response{
		Type:  Replace,
		Value: val,
	}, nil
}

func init() {
	RegisterOp("stringify", StringifyOperator{})
}
