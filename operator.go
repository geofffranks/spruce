package main

import (
	"regexp"
	"fmt"
)

type Action int

const (
	Replace Action = iota
	Inject
)

type Response struct {
	Type  Action
	Value interface{}
}

type Operator interface {
	// evaluate the tree and determine what should be done to satisfy caller
	Run(ev *Evaluator, args []interface{}) (*Response, error)

	// returns a set of implicit / inherent dependencies used by Run()
	Dependencies(ev *Evaluator, args []interface{}, locs []*Cursor) []*Cursor
}

var OpRegistry map[string]Operator

func OperatorFor(name string) Operator {
	if op, ok := OpRegistry[name]; ok {
		return op
	}
	return NullOperator{Missing: name}
}

func RegisterOp(name string, op Operator) {
	if OpRegistry == nil {
		OpRegistry = map[string]Operator{}
	}
	OpRegistry[name] = op
}

type Opcall struct {
	src   string
	where *Cursor
	canonical *Cursor
	op    Operator
	args  []interface{}
}

func ParseOpcall(src string) (*Opcall, error) {
	split := func(src string) []string {
		list := make([]string, 0, 0)

		buf := ""
		escaped := false
		quoted := false

		for _, c := range src {
			if escaped {
				buf += string(c)
				escaped = false
				continue
			}

			if c == '\\' {
				escaped = true
				continue
			}

			if c == ' ' || c == '\t' || c == ',' {
				if quoted {
					buf += string(c)
					continue
				} else if buf != "" {
					list = append(list, buf)
					buf = ""
				}
				continue
			}

			if c == '"' {
				buf += string(c)
				quoted = !quoted
				continue
			}

			buf += string(c)
		}

		if buf != "" {
			list = append(list, buf)
		}

		return list
	}

	argify := func(src string) ([]interface{}, error) {
		var args []interface{}

		quoted_string := regexp.MustCompile(`^"(.*)"$`)
		numeric       := regexp.MustCompile(`^\d+$`)

		for i, arg := range split(src) {
			switch {
			case quoted_string.MatchString(arg):
				m := quoted_string.FindStringSubmatch(arg)
				DEBUG("  #%d: parsed as quoted string literal '%s'", i, m[1])
				args = append(args, m[1])

			case numeric.MatchString(arg):
				DEBUG("  #%d: parsed as unquoted integer literal '%s'", i, arg)
				args = append(args, arg)

			default:
				c, err := ParseCursor(arg)
				if err != nil {
					DEBUG("  #%d: %s is a malformed reference: %s", i, arg, err)
					return args, err
				}
				DEBUG("  #%d: parsed as a reference to $.%s", i, c)
				args = append(args, c)
			}
		}
		DEBUG("")

		return args, nil
	}

	op := &Opcall{ src: src }

	for _, pattern := range []string{
		`^\Q((\E\s*([a-zA-Z][a-zA-Z0-9_-]*)(?:\s*\((.*)\))?\s*\Q))\E$`, // (( op(x,y,z) ))
		`^\Q((\E\s*([a-zA-Z][a-zA-Z0-9_-]*)(?:\s+(.*))?\s*\Q))\E$`,     // (( op x y z ))
	} {
		re := regexp.MustCompile(pattern)
		if !re.MatchString(src) {
			continue
		}

		m := re.FindStringSubmatch(src)
		DEBUG("parsing `%s': looks like a (( %s ... )) operator\n arguments:", src, m[1])
		args, err := argify(m[2])
		if err != nil {
			return nil, err
		}
		if len(args) == 0 {
			DEBUG("  (none)")
		}
		op.op = OperatorFor(m[1])
		op.args = args
		return op, nil
	}

	return nil, nil
}

func (op *Opcall) Dependencies(ev *Evaluator, locs []*Cursor) []*Cursor {
	l := []*Cursor{}
	for _, arg := range op.args {
		if cursor, ok := arg.(*Cursor); ok {
			c := cursor.Copy()
			for c.Depth() > 0 {
				if _, err := c.Canonical(ev.Tree); err == nil {
					break
				}
				c.Pop()
			}
			if c.Depth() == 0 {
				continue
			}
			canon, _ := c.Canonical(ev.Tree)
			l = append(l, canon)
		}
	}

	for _, c := range op.op.Dependencies(ev, op.args, locs) {
		l = append(l, c)
	}
	return l
}

func (op *Opcall) Run(ev *Evaluator) (*Response, error) {
	was := ev.Here
	ev.Here = op.where
	r, err := op.op.Run(ev, op.args)
	ev.Here = was

	if err != nil {
		return nil, fmt.Errorf("$.%s: %s", op.where, err)
	}
	return r, nil
}
