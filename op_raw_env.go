package spruce

import (
	"fmt"
	"os"

	"github.com/starkandwayne/goutils/tree"

	. "github.com/geofffranks/spruce/log"
)

// ResolveRawEnv resolves an expression like Resolve, but keeps environment
// variable values as raw strings without YAML unmarshaling.
func (e *Expr) ResolveRawEnv(tree map[interface{}]interface{}) (*Expr, error) {
	switch e.Type {
	case EnvVar:
		v, ok := os.LookupEnv(e.Name)
		if !ok {
			return nil, fmt.Errorf("environment variable $%s is not set", e.Name)
		}
		return &Expr{Type: Literal, Literal: v}, nil

	case LogicalOr:
		if o, err := e.Left.ResolveRawEnv(tree); err == nil {
			return o, nil
		}
		return e.Right.ResolveRawEnv(tree)

	default:
		return e.Resolve(tree)
	}
}

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
		return nil, fmt.Errorf("raw_env operator requires exactly one argument")
	}

	// validate that the leftmost leaf is an environment variable
	leftmost := args[0]
	for leftmost.Type == LogicalOr {
		leftmost = leftmost.Left
	}
	if leftmost.Type != EnvVar {
		return nil, fmt.Errorf("raw_env operator only accepts environment variable arguments")
	}

	v, err := args[0].ResolveRawEnv(ev.Tree)
	if err != nil {
		DEBUG("  %s", err)
		return nil, err
	}

	switch v.Type {
	case Literal:
		return &Response{Type: Replace, Value: v.Literal}, nil
	case Reference:
		val, err := v.Reference.Resolve(ev.Tree)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve `%s`: %s", v.Reference, err)
		}
		return &Response{Type: Replace, Value: val}, nil
	default:
		return nil, fmt.Errorf("raw_env operator received unexpected expression type")
	}
}

func init() {
	RegisterOp("raw_env", RawEnvOperator{})
}
