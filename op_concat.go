package main

import (
	"fmt"
	"strings"
)

type ConcatOperator struct {}

func (ConcatOperator) Dependencies(_ *Evaluator, _ []interface{}, _ []*Cursor) []*Cursor {
	return []*Cursor{}
}

func (ConcatOperator) Run(ev *Evaluator, args []interface{}) (*Response, error) {
	DEBUG("running (( concat ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( concat ... )) operation at $%s\n", ev.Here)

	var l []string

	if len(args) < 2 {
		return nil, fmt.Errorf("concat operator requires at least two arguments")
	}

	for i, arg := range args {
		switch arg.(type) {
		case string:
			DEBUG("  arg[%d]: using string literal '%s'", i, arg.(string))
			DEBUG("     [%d]: appending '%s' to resultant string", i, arg.(string))
			l = append(l, arg.(string))

		case *Cursor:
			c := arg.(*Cursor)
			DEBUG("  arg[%d]: trying to resolve reference $.%s", i, c.String())
			s, err := c.Resolve(ev.Tree)
			if err != nil {
				DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
				return nil, fmt.Errorf("Unable to resolve `%s`: %s", c, err)
			}

			var v string
			switch s.(type) {
			case map[interface{}]interface{}:
				DEBUG("  arg[%d]: %v is not a string scalar", i, s)
				return nil, fmt.Errorf("tried to concat %s, which is not a string scalar", c)

			case []interface{}:
				DEBUG("  arg[%d]: %v is not a string scalar", i, s)
				return nil, fmt.Errorf("tried to concat %s, which is not a string scalar", c)

			default:
				v = fmt.Sprintf("%v", s)
			}
			DEBUG("     [%d]: appending '%s' to resultant string", i, s)
			l = append(l, v)

		default:
			DEBUG("  arg[%d]: I don't know what to do with '%v'", i, arg)
			return nil, fmt.Errorf("concat operator only accepts string literals and key reference arguments")
		}
		DEBUG("")
	}

	final := strings.Join(l, "")
	DEBUG("  resolved (( concat ... )) operation to the string:\n    \"%s\"", final)

	return &Response{
		Type: Replace,
		Value: final,
	}, nil
}

func init() {
	RegisterOp("concat", ConcatOperator{})
}
