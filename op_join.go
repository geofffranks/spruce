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

	if len(args) == 1 {
		DEBUG("  too few arguments supplied to (( join ... )) operation.")
		return nil, ansi.Errorf("too few arguments supplied to @c{(( join ... ))}")
	}

	var seperator string
	var list []string

	for i, arg := range args {
		if i == 0 { // argument #0: seperator
			sep, err := arg.Resolve(ev.Tree)
			if err != nil {
				DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
				return nil, err
			}

			if sep.Type != Literal {
				DEBUG("     [%d]: unsupported type for join operator seperator argument: '%v'", i, sep)
				return nil, fmt.Errorf("join operator only accepts literal argument for the seperator")
			}

			DEBUG("     [%d]: list seperator will be: %s", i, sep)
			seperator = sep.Literal.(string)

		} else { // argument #1..n: list, or literal
			ref, err := arg.Resolve(ev.Tree)
			if err != nil {
				DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
				return nil, err
			}

			switch ref.Type {
			case Literal:
				DEBUG("     [%d]: adding literal %s to the list", i, ref)
				list = append(list, fmt.Sprintf("%v", ref.Literal))

			case Reference:
				DEBUG("     [%d]: trying to resolve reference $.%s", i, ref.Reference)
				s, err := ref.Reference.Resolve(ev.Tree)
				if err != nil {
					DEBUG("     [%d]: resolution failed with error: %s", i, err)
					return nil, fmt.Errorf("Unable to resolve `%s`: %s", ref.Reference, err)
				}

				switch s.(type) {
				case []interface{}:
					DEBUG("     [%d]: $.%s is a list", i, ref.Reference)
					for idx, entry := range s.([]interface{}) {
						switch entry.(type) {
						case []interface{}:
							DEBUG("     [%d]: entry #%d in list is a list (not a literal)", i, idx)
							return nil, ansi.Errorf("entry #%d in list is not compatible for @c{(( join ... ))}", idx)

						case map[interface{}]interface{}:
							DEBUG("     [%d]: entry #%d in list is a map (not a literal)", i, idx)
							return nil, ansi.Errorf("entry #%d in list is not compatible for @c{(( join ... ))}", idx)

						default:
							list = append(list, fmt.Sprintf("%v", entry))
						}
					}

				case map[interface{}]interface{}:
					DEBUG("     [%d]: $.%s is a map (not a list or a literal)", i, ref.Reference)
					return nil, ansi.Errorf("referenced entry is not a list or string for @c{(( join ... ))}")

				default:
					DEBUG("     [%d]: $.%s is a literal", i, ref.Reference)
					list = append(list, fmt.Sprintf("%v", s))
				}

			default:
				DEBUG("     [%d]: unsupported type for join operator: '%v'", i, ref)
				return nil, fmt.Errorf("join operator only lists with string entries, and literals as data arguments")
			}
		}
	}

	// finally, join and return
	DEBUG("  joined list: %s", strings.Join(list, seperator))
	return &Response{
		Type:  Replace,
		Value: strings.Join(list, seperator),
	}, nil
}

func init() {
	RegisterOp("join", JoinOperator{})
}
