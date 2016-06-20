package spruce

import (
	"fmt"
	"strings"

	"github.com/jhunt/ansi"
	"github.com/jhunt/tree"

	. "github.com/geofffranks/spruce/log"
)

// JoinOperator ...
type JoinOperator struct{}

// Setup ...
func (JoinOperator) Setup() error {
	return nil
}

// Phase ...
func (JoinOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (JoinOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor) []*tree.Cursor {
	return []*tree.Cursor{}
}

// Run ...
func (JoinOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( join ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( join ... )) operation at $%s\n", ev.Here)

	if len(args) == 0 {
		DEBUG("  no arguments supplied to (( join ... )) operation.")
		return nil, ansi.Errorf("no arguments specified to @c{(( join ... ))}")
	}

	if len(args) > 2 {
		DEBUG("  too many arguments supplied to (( join ... )) operation.")
		return nil, ansi.Errorf("too many arguments supplied to @c{(( join ... ))}")
	}

	// --- argument #0: list ---
	var list []string

	ref, err := args[0].Resolve(ev.Tree)
	if err != nil {
		DEBUG("     [list]: resolution failed\n    error: %s", err)
		return nil, err
	}

	if ref.Type != Reference {
		DEBUG("     [list]: unsupported type for join operator list argument: '%v'", ref)
		return nil, fmt.Errorf("join operator only accepts reference argument for the list to be joined")
	}

	DEBUG("     [list]: trying to resolve reference $.%s", ref.Reference)
	s, err := ref.Reference.Resolve(ev.Tree)
	if err != nil {
		DEBUG("     [list]: resolution failed with error: %s", err)
		return nil, fmt.Errorf("Unable to resolve `%s`: %s", ref.Reference, err)
	}

	switch s.(type) {
	case []interface{}:
		DEBUG("     [list]: $.%s is a list; good", ref.Reference)
		for idx, entry := range s.([]interface{}) {
			if str, ok := entry.(string); ok {
				list = append(list, str)

			} else {
				DEBUG("     [list]: entry #%d in list is not a string", idx)
				return nil, ansi.Errorf("entry #%d in list is not compatible for @c{(( join ... ))}", idx)
			}
		}

	default:
		DEBUG("     [list]: $.%s is not a list", ref.Reference)
		return nil, ansi.Errorf("referenced argument is not a list for @c{(( join ... ))}")
	}

	// --- argument #1: seperator ---
	var seperator string

	sep, err := args[1].Resolve(ev.Tree)
	if err != nil {
		DEBUG("     [seperator]: resolution failed\n    error: %s", err)
		return nil, err
	}

	if sep.Type != Literal {
		DEBUG("     [seperator]: unsupported type for join operator seperator argument: '%v'", ref)
		return nil, fmt.Errorf("join operator only accepts literal argument for the seperator")
	}

	DEBUG("     [seperator]: list seperator will be: %s", sep)
	seperator = sep.Literal.(string)

	// --- join ---

	DEBUG("  joined list: %s", strings.Join(list, seperator))
	return &Response{
		Type:  Replace,
		Value: strings.Join(list, seperator),
	}, nil
}

func init() {
	RegisterOp("join", JoinOperator{})
}
