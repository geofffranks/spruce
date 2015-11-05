package main

import (
	"fmt"
)

type GrabOperator struct {}

func (GrabOperator) Dependencies(_ *Evaluator, _ []interface{}, _ []*Cursor) []*Cursor {
	return []*Cursor{}
}

func (GrabOperator) Run(ev *Evaluator, args []interface{}) (*Response, error) {
	DEBUG("running (( grab ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( grab ... )) operation at $%s\n", ev.Here)

	var vals []interface{}

	for i, arg := range args {
		switch arg.(type) {
		case string:
			DEBUG("  arg[%d]: found string literal '%s'", i, arg.(string))
			DEBUG("           (grab operator only handles references to other parts of the YAML tree)")
			return nil, fmt.Errorf("grab operator only accepts key reference arguments")

		case *Cursor:
			c := arg.(*Cursor)
			DEBUG("  arg[%d]: trying to resolve reference $.%s", i, c.String())
			v, err := c.Resolve(ev.Tree)
			if err != nil {
				DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
				return nil, fmt.Errorf("Unable to resolve `%s`: %s", c, err)
			}
			DEBUG("     [%d]: resolved to a value (could be a map, a list or a scalar); appending", i)
			vals = append(vals, v)

		default:
			DEBUG("  arg[%d]: I don't know what to do with '%v'", i, arg)
			return nil, fmt.Errorf("grab operator only accepts key reference arguments")
		}
		DEBUG("")
	}

	switch len(args) {
	case 0:
		DEBUG("  no arguments supplied to (( grab ... )) operation.  oops.")
		return nil, fmt.Errorf("no arguments specified to (( grab ... ))")

	case 1:
		DEBUG("  called with only one argument; returning value as-is")
		return &Response{
			Type: Replace,
			Value: vals[0],
		}, nil

	default:
		DEBUG("  called with more than one arguments; flattening top-level lists into a single list")
		flat := []interface{}{}
		for i, lst := range vals {
			switch lst.(type) {
			case []interface{}:
				DEBUG("    [%d]: $.%s is a list; flattening it out", i, args[i].(*Cursor))
				flat = append(flat, lst.([]interface{})...)
			default:
				DEBUG("    [%d]: $.%s is not a list; appending it as-is", i, args[i].(*Cursor))
				flat = append(flat, lst)
			}
		}
		DEBUG("")

		return &Response{
			Type: Replace,
			Value: flat,
		}, nil
	}
}

func init() {
	RegisterOp("grab", GrabOperator{})
}
