package spruce

import (
	"fmt"
	"strings"

	"github.com/starkandwayne/goutils/tree"

	. "github.com/geofffranks/spruce/log"
)

// DeferOperator sheds the "defer" command off of (( defer args args args )) and
// leaves (( args args args ))
type DeferOperator struct{}

// Setup doesn't do anything for Defer. We're a pretty lightweight operator.
func (DeferOperator) Setup() error {
	return nil
}

// Phase gives back Param phase in this case, because we don't want any
// following phases to pick up the operator post-deference
func (DeferOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies returns an empty slice - defer produces no deps at all.
func (DeferOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor, _ []*tree.Cursor) []*tree.Cursor {
	return nil
}

// Run chops off "defer" and leaves the args in double parens. Need to
// reconstruct the operator string
func (DeferOperator) Run(_ *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("Running defer operator...")
	if len(args) == 0 {
		return nil, fmt.Errorf("Defer has no arguments - what are you deferring?")
	}

	components := []string{"(("} //Join these with spaces at the end

	for _, arg := range args {
		components = append(components, arg.String())
	}
	components = append(components, "))")

	DEBUG("Returning from defer operator")

	return &Response{
		Type:  Replace,
		Value: strings.Join(components, " "),
	}, nil
}

func init() {
	RegisterOp("defer", DeferOperator{})
}
