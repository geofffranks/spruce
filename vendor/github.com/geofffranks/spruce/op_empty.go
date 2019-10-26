package spruce

import (
	"fmt"
	"strings"

	"github.com/starkandwayne/goutils/tree"
)

// EmptyOperator allows the user to emplace an empty array, hash, or string into
// the YAML datastructure.
type EmptyOperator struct{}

// Setup ...
func (EmptyOperator) Setup() error {
	return nil
}

// Phase ...
func (EmptyOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (EmptyOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor, _ []*tree.Cursor) []*tree.Cursor {
	return nil
}

// Run ...
func (EmptyOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("empty operator expects 1 argument, received %d", len(args))
	}

	var emptyType string

	switch args[0].Type {
	case Literal:
		var isString bool
		emptyType, isString = args[0].Literal.(string)
		if !isString {
			return nil, fmt.Errorf("cannot interpret argument for empty operator")
		}
	case Reference:
		emptyType = strings.TrimPrefix(args[0].Reference.String(), ".")
	default:
		return nil, fmt.Errorf("cannot interpret argument for empty operator")
	}

	var value interface{}
	switch emptyType {
	case "hash", "map":
		value = map[string]interface{}{}
	case "array", "list":
		value = []interface{}{}
	case "string":
		value = ""
	default:
		return nil, fmt.Errorf("unknown type for empty operator: %s", emptyType)
	}
	return &Response{
		Type:  Replace,
		Value: value,
	}, nil
}

func init() {
	RegisterOp("empty", EmptyOperator{})
}
