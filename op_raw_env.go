package spruce

import (
	"fmt"
	"os"

	"github.com/starkandwayne/goutils/tree"

	. "github.com/geofffranks/spruce/log"
)

// RawEnvOperator retrieves an environment variable as a raw string without YAML unmarshaling
type RawEnvOperator struct{}

// Setup ...
func (RawEnvOperator) Setup() error {
	return nil
}

// Phase ...
func (RawEnvOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (RawEnvOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor, auto []*tree.Cursor) []*tree.Cursor {
	return auto
}

// Run ...
func (RawEnvOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( raw_env ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( raw_env ... )) operation at $%s\n", ev.Here)

	if len(args) != 1 {
		return nil, fmt.Errorf("raw_env operator requires exactly one reference argument")
	}

	arg := args[0]

	if arg.Type != EnvVar {
		DEBUG(" arg not an environment variable reference")
		return nil, fmt.Errorf("raw_env operator only accepts environment variable arguments")
	}

	DEBUG(" arg: retrieving environment variable $%s", arg.Name)
	v := os.Getenv(arg.Name)
	if v == "" {
		DEBUG("  environment variable $%s is not set", arg.Name)
		return nil, fmt.Errorf("environment variable $%s is not set", arg.Name)
	}

	DEBUG("  resolved to raw string value (length: %d)", len(v))
	DEBUG("")
	return &Response{
		Type:  Replace,
		Value: v,
	}, nil
}

func init() {
	RegisterOp("raw_env", RawEnvOperator{})
}
