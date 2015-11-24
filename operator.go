package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

// Action ...
type Action int

const (
	// Replace ...
	Replace Action = iota

	// Inject ...
	Inject
)

// OperatorPhase ...
type OperatorPhase int

const (
	// MergePhase ...
	MergePhase OperatorPhase = iota
	// EvalPhase ...
	EvalPhase
	// CheckPhase ...
	CheckPhase
)

// Response ...
type Response struct {
	Type  Action
	Value interface{}
}

// Operator ...
type Operator interface {
	// setup whatever global/static state needed -- see (( static_ips ... ))
	Setup() error

	// evaluate the tree and determine what should be done to satisfy caller
	Run(ev *Evaluator, args []*Expr) (*Response, error)

	// returns a set of implicit / inherent dependencies used by Run()
	Dependencies(ev *Evaluator, args []*Expr, locs []*Cursor) []*Cursor

	// what phase does this operator run during?
	Phase() OperatorPhase
}

// OpRegistry ...
var OpRegistry map[string]Operator

// OperatorFor ...
func OperatorFor(name string) Operator {
	if op, ok := OpRegistry[name]; ok {
		return op
	}
	return NullOperator{Missing: name}
}

// RegisterOp ...
func RegisterOp(name string, op Operator) {
	if OpRegistry == nil {
		OpRegistry = map[string]Operator{}
	}
	OpRegistry[name] = op
}

// SetupOperators ...
func SetupOperators(phase OperatorPhase) error {
	errors := MultiError{Errors: []error{}}
	for _, op := range OpRegistry {
		if op.Phase() == phase {
			if err := op.Setup(); err != nil {
				errors.Append(err)
			}
		}
	}
	if len(errors.Errors) > 0 {
		return errors
	}
	return nil
}

// ExprType ...
type ExprType int

const (
	// Reference ...
	Reference ExprType = iota
	// Literal ...
	Literal
	// LogicalOr ...
	LogicalOr
)

// Expr ...
type Expr struct {
	Type      ExprType
	Reference *Cursor
	Literal   interface{}
	Left      *Expr
	Right     *Expr
}

func (e *Expr) String() string {
	switch e.Type {
	case Literal:
		if e.Literal == nil {
			return "nil"
		}
		if _, ok := e.Literal.(string); ok {
			return fmt.Sprintf(`"%s"`, e.Literal)
		}
		return fmt.Sprintf("%v", e.Literal)

	case Reference:
		return e.Reference.String()

	case LogicalOr:
		return fmt.Sprintf("(%s || %s)", e.Left, e.Right)

	default:
		return "<!! unknown !!>"
	}
}

// Reduce ...
func (e *Expr) Reduce() (*Expr, error) {

	var reduce func(*Expr) (*Expr, *Expr, bool)
	reduce = func(e *Expr) (*Expr, *Expr, bool) {
		switch e.Type {
		case Literal:
			return e, e, false
		case Reference:
			return e, nil, false

		case LogicalOr:
			l, short, _ := reduce(e.Left)
			if short != nil {
				return l, short, true
			}

			r, short, more := reduce(e.Right)
			return &Expr{
				Type:  LogicalOr,
				Left:  l,
				Right: r,
			}, short, more
		}
		return nil, nil, false
	}

	reduced, short, more := reduce(e)
	if more && short != nil {
		return reduced, fmt.Errorf("literal %v short-circuits expression", short)
	}
	return reduced, nil
}

// Resolve ...
func (e *Expr) Resolve(tree map[interface{}]interface{}) (*Expr, error) {
	switch e.Type {
	case Literal:
		return e, nil

	case Reference:
		if _, err := e.Reference.Resolve(tree); err != nil {
			return nil, fmt.Errorf("Unable to resolve `%s`: %s", e.Reference, err)
		}
		return e, nil

	case LogicalOr:
		if o, err := e.Left.Resolve(tree); err == nil {
			return o, nil
		}
		return e.Right.Resolve(tree)
	}
	return nil, fmt.Errorf("unknown expression operand type (%d)", e.Type)
}

// Evaluate ...
func (e *Expr) Evaluate(tree map[interface{}]interface{}) (interface{}, error) {
	final, err := e.Resolve(tree)
	if err != nil {
		return nil, err
	}

	switch final.Type {
	case Literal:
		return final.Literal, nil
	case Reference:
		return final.Reference.Resolve(tree)
	case LogicalOr:
		return nil, fmt.Errorf("expression resolved to a logical OR operation (which shouldn't happen)")
	}
	return nil, fmt.Errorf("unknown operand type")
}

// Dependencies ...
func (e *Expr) Dependencies(ev *Evaluator, locs []*Cursor) []*Cursor {
	l := []*Cursor{}

	canonicalize := func(c *Cursor) {
		cc := c.Copy()
		for cc.Depth() > 0 {
			if _, err := cc.Canonical(ev.Tree); err == nil {
				break
			}
			cc.Pop()
		}
		if cc.Depth() > 0 {
			l = append(l, cc)
		}
	}

	switch e.Type {
	case Reference:
		canonicalize(e.Reference)

	case LogicalOr:
		for _, c := range e.Left.Dependencies(ev, locs) {
			canonicalize(c)
		}
		for _, c := range e.Right.Dependencies(ev, locs) {
			canonicalize(c)
		}
	}

	return l
}

// Opcall ...
type Opcall struct {
	src       string
	where     *Cursor
	canonical *Cursor
	op        Operator
	args      []*Expr
}

// ParseOpcall ...
func ParseOpcall(phase OperatorPhase, src string) (*Opcall, error) {
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

	argify := func(src string) (args []*Expr, err error) {
		qstring := regexp.MustCompile(`^"(.*)"$`)
		integer := regexp.MustCompile(`^[+-]?\d+(\.\d+)?$`)
		float := regexp.MustCompile(`^[+-]?\d*\.\d+$`)

		var stack []*Expr
		for i, arg := range split(src) {
			switch {
			case qstring.MatchString(arg):
				m := qstring.FindStringSubmatch(arg)
				DEBUG("  #%d: parsed as quoted string literal '%s'", i, m[1])
				stack = append(stack, &Expr{Type: Literal, Literal: m[1]})

			case float.MatchString(arg):
				DEBUG("  #%d: parsed as unquoted floating point literal '%s'", i, arg)
				v, err := strconv.ParseFloat(arg, 64)
				if err != nil {
					DEBUG("  #%d: %s is not parseable as a floatin point number: %s", i, arg, err)
					return args, err
				}
				stack = append(stack, &Expr{Type: Literal, Literal: v})

			case integer.MatchString(arg):
				DEBUG("  #%d: parsed as unquoted integer literal '%s'", i, arg)
				v, err := strconv.ParseInt(arg, 10, 64)
				if err != nil {
					DEBUG("  #%d: %s is not parseable as an integer: %s", i, arg, err)
					return args, err
				}
				stack = append(stack, &Expr{Type: Literal, Literal: v})

			case arg == "||":
				DEBUG("  #%d: parsed logical-or operator, '||'", i)
				stack = append(stack, &Expr{Type: LogicalOr})

			case arg == "nil" || arg == "null" || arg == "~":
				DEBUG("  #%d: parsed the nil value token '%s'", i, arg)
				stack = append(stack, &Expr{Type: Literal, Literal: nil})

			default:
				c, err := ParseCursor(arg)
				if err != nil {
					DEBUG("  #%d: %s is a malformed reference: %s", i, arg, err)
					return args, err
				}
				DEBUG("  #%d: parsed as a reference to $.%s", i, c)
				stack = append(stack, &Expr{Type: Reference, Reference: c})
			}
		}
		DEBUG("")

		push := func(e *Expr) {
			if e == nil {
				return
			}
			reduced, err := e.Reduce()
			if err != nil {
				fmt.Fprintf(os.Stdout, "warning: %s\n", err)
			}
			args = append(args, reduced)
		}

		var e *Expr
		for len(stack) > 0 {
			if e == nil {
				e = stack[0]
				stack = stack[1:]
				continue
			}

			if stack[0].Type == LogicalOr {
				stack[0].Left = e
				e = stack[0]
				stack = stack[1:]
				continue
			}

			if e.Type == LogicalOr {
				if e.Right != nil {
					e = &Expr{Type: LogicalOr, Left: e}
				}
				e.Right = stack[0]
				stack = stack[1:]
				continue
			}

			push(e)
			e = stack[0]
			stack = stack[1:]
		}
		push(e)

		return args, nil
	}

	op := &Opcall{src: src}

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

		op.op = OperatorFor(m[1])
		if op.op.Phase() != phase {
			DEBUG("  - skipping (( %s ... )) operation; it belongs to a different phase", m[1])
			return nil, nil
		}

		args, err := argify(m[2])
		if err != nil {
			return nil, err
		}
		if len(args) == 0 {
			DEBUG("  (none)")
		}
		op.args = args
		return op, nil
	}

	return nil, nil
}

// Dependencies ...
func (op *Opcall) Dependencies(ev *Evaluator, locs []*Cursor) []*Cursor {
	l := []*Cursor{}
	for _, arg := range op.args {
		for _, c := range arg.Dependencies(ev, locs) {
			l = append(l, c)
		}
	}

	for _, c := range op.op.Dependencies(ev, op.args, locs) {
		l = append(l, c)
	}
	return l
}

// Run ...
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
