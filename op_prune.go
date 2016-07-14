package spruce

import (
	"fmt"

	"github.com/starkandwayne/goutils/tree"

	. "github.com/geofffranks/spruce/log"
)

var keysToPrune []string

func addToPruneListIfNecessary(paths ...string) {
	for _, path := range paths {
		if !isIncluded(keysToPrune, path) {
			DEBUG("adding '%s' to the list of paths to prune", path)
			keysToPrune = append(keysToPrune, path)
		}
	}
}

func isIncluded(list []string, name string) bool {
	for _, entry := range list {
		if entry == name {
			return true
		}
	}

	return false
}

// PruneOperator ...
type PruneOperator struct{}

// Setup ...
func (PruneOperator) Setup() error {
	return nil
}

// Phase ...
func (PruneOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (PruneOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor) []*tree.Cursor {
	return []*tree.Cursor{}
}

// Run ...
func (PruneOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( prune ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( prune ... )) operation at $.%s\n", ev.Here)

	addToPruneListIfNecessary(fmt.Sprintf("%s", ev.Here))

	// simply replace it with nil (will be pruned at the end anyway)
	return &Response{
		Type:  Replace,
		Value: nil,
	}, nil
}

func init() {
	RegisterOp("prune", PruneOperator{})
}
