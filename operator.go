// Package spruce provides a YAML merging tool with support for operators,
// references, and various merge strategies.
package spruce

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/geofffranks/yaml"
	"github.com/starkandwayne/goutils/ansi"

	. "github.com/geofffranks/spruce/log"
	"github.com/starkandwayne/goutils/tree"
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
	// ParamPhase ...
	ParamPhase
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
	Dependencies(ev *Evaluator, args []*Expr, locs []*tree.Cursor, auto []*tree.Cursor) []*tree.Cursor

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
	EnvVar
)

// Expr ...
type Expr struct {
	Type           ExprType
	Reference      *tree.Cursor
	BracketedNodes []bool
	Literal        interface{}
	Name           string
	Left           *Expr
	Right          *Expr
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

	case EnvVar:
		return fmt.Sprintf("$%s", e.Name)

	case Reference:
		return e.Reference.String()

	case LogicalOr:
		return fmt.Sprintf("%s || %s", e.Left, e.Right)

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
		case EnvVar:
			return e, nil, false
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
		return reduced, NewWarningError(eContextAll, "@R{literal} @c{%v} @R{short-circuits expression (}@c{%s}@R{)}", short, e)
	}
	return reduced, nil
}

// Resolve ...
func (e *Expr) Resolve(tree map[interface{}]interface{}) (*Expr, error) {
	switch e.Type {
	case Literal:
		return e, nil

	case EnvVar:
		v := os.Getenv(e.Name)
		if v == "" {
			return nil, ansi.Errorf("@R{Environment variable} @c{$%s} @R{is not set}", e.Name)
		}

		var val interface{}
		err := yaml.Unmarshal([]byte(v), &val)
		_, isString := val.(string)
		if isString || err != nil {
			return &Expr{Type: Literal, Literal: v}, nil
		}
		return &Expr{Type: Literal, Literal: val}, nil

	case Reference:
		// Clear bracket flags for $VAR nodes before ResolveEnv consumes the "$" sentinel,
		// so (( grab map[$VAR] )) treats the env-var value as a literal key rather than
		// as a YAML path to re-resolve against the tree.
		for i, node := range e.Reference.Nodes {
			if i < len(e.BracketedNodes) && len(node) > 0 && node[0] == '$' {
				e.BracketedNodes[i] = false
			}
		}
		e.Reference.Nodes = ResolveEnv(e.Reference.Nodes)
		nodes, err := ResolveDynamicRefs(e.Reference.Nodes, e.BracketedNodes, tree)
		if err != nil {
			return nil, ansi.Errorf("@R{%s}", err)
		}
		e.Reference.Nodes = nodes
		if _, err := e.Reference.Resolve(tree); err != nil {
			return nil, ansi.Errorf("@R{unable to resolve `}@c{%s}@R{`: %s}", e.Reference, err)
		}
		return e, nil

	case LogicalOr:
		if o, err := e.Left.Resolve(tree); err == nil {
			return o, nil
		}
		return e.Right.Resolve(tree)
	}
	return nil, ansi.Errorf("@R{unknown expression operand type (}@c{%d}@R{)}", e.Type)
}

func ResolveEnv(nodes []string) []string {
	var resolved []string
	for _, node := range nodes {
		if len(node) > 0 && node[0] == '$' {
			resolved = append(resolved, os.Getenv(node[1:]))
		} else {
			resolved = append(resolved, node)
		}
	}
	return resolved
}

// bracketsOf re-parses a path string to recover which nodes were specified in
// bracket notation. tree.ParseCursor normalizes both notations into identical
// Nodes, discarding this distinction, so we must re-derive it from the source.
func bracketsOf(path string) []bool {
	var result []bool
	inBracket := false
	var buf bytes.Buffer

	push := func(isBracketed bool) {
		if buf.Len() == 0 {
			return
		}
		// Mirror ParseCursor: skip leading "$" sentinel
		if len(result) == 0 && buf.String() == "$" {
			buf.Reset()
			return
		}
		result = append(result, isBracketed)
		buf.Reset()
	}

	for _, c := range path {
		switch c {
		case '.':
			if inBracket {
				buf.WriteRune(c)
			} else {
				push(false)
			}
		case '[':
			push(false)
			inBracket = true
		case ']':
			push(true)
			inBracket = false
		default:
			buf.WriteRune(c)
		}
	}
	push(false)
	return result
}

// ResolveDynamicRefs resolves nodes that were specified in bracket notation as YAML
// path references against the tree, enabling (( grab map[key_ref] )) syntax.
// Mutates bracketed: clears flags for nodes that were substituted, to keep
// subsequent calls idempotent (e.g. when Dependencies and Run both invoke Resolve).
func ResolveDynamicRefs(nodes []string, bracketed []bool, t map[interface{}]interface{}) ([]string, error) {
	hasDynamic := false
	for _, b := range bracketed {
		if b {
			hasDynamic = true
			break
		}
	}
	if !hasDynamic {
		return nodes, nil
	}

	resolved := make([]string, len(nodes))
	copy(resolved, nodes)
	for i, node := range nodes {
		if i >= len(bracketed) || !bracketed[i] {
			continue
		}
		if len(node) == 0 || node[0] == '$' {
			continue // $VAR nodes were already substituted by ResolveEnv; skip any remnants
		}
		if _, err := strconv.ParseUint(node, 10, 64); err == nil {
			continue // numeric nodes are array indices, not YAML references
		}
		c, err := tree.ParseCursor(node)
		if err != nil {
			return nil, fmt.Errorf("invalid bracketed key reference %q: %s", node, err)
		}
		val, err := c.Resolve(t)
		if err != nil {
			return nil, fmt.Errorf("unable to resolve bracketed key reference %q: %s", node, err)
		}
		switch v := val.(type) {
		case string:
			resolved[i] = v
		case int:
			resolved[i] = strconv.Itoa(v)
		case int64:
			resolved[i] = strconv.FormatInt(v, 10)
		case float64:
			resolved[i] = strconv.FormatFloat(v, 'f', -1, 64)
		case bool:
			resolved[i] = strconv.FormatBool(v)
		default:
			return nil, fmt.Errorf("bracketed key reference %q resolved to %T; expected a scalar (string, int, float, or bool)", node, val)
		}
		// Clear the flag so re-resolving the same Expr (e.g. in JoinOperator.Dependencies
		// followed by Run) treats the already-substituted node as a literal key.
		bracketed[i] = false
	}
	return resolved, nil
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
	case EnvVar:
		return os.Getenv(final.Name), nil
	case Reference:
		return final.Reference.Resolve(tree)
	case LogicalOr:
		return nil, fmt.Errorf("expression resolved to a logical OR operation (which shouldn't happen)")
	}
	return nil, fmt.Errorf("unknown operand type")
}

// Dependencies ...
func (e *Expr) Dependencies(ev *Evaluator, locs []*tree.Cursor) []*tree.Cursor {
	l := []*tree.Cursor{}

	canonicalize := func(c *tree.Cursor) {
		cc := c.Copy()
		for cc.Depth() > 0 {
			if _, err := cc.Canonical(ev.Tree); err == nil {
				break
			}
			cc.Pop()
		}
		if cc.Depth() > 0 {
			canon, _ := cc.Canonical(ev.Tree)
			l = append(l, canon)
		}
	}

	switch e.Type {
	case Reference:
		canonicalize(e.Reference)
		for i, node := range e.Reference.Nodes {
			if i >= len(e.BracketedNodes) || !e.BracketedNodes[i] {
				continue
			}
			if len(node) == 0 || node[0] == '$' {
				continue
			}
			if _, err := strconv.ParseUint(node, 10, 64); err == nil {
				continue
			}
			if c, err := tree.ParseCursor(node); err == nil {
				canonicalize(c)
			}
		}

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
	where     *tree.Cursor
	canonical *tree.Cursor
	op        Operator
	args      []*Expr
}

// ParseOpcall ...
func ParseOpcall(phase OperatorPhase, src string) (*Opcall, error) {
	split := func(src string) []string {
		list := make([]string, 0)

		buf := ""
		escaped := false
		quoted := false

		for _, c := range src {
			if escaped {
				switch c {
				case 'n':
					buf += "\n"
				case 'r':
					buf += "\r"
				case 't':
					buf += "\t"
				default:
					buf += string(c)
				}
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
				} else {
					if buf != "" {
						list = append(list, buf)
						buf = ""
					}
					if c == ',' {
						list = append(list, ",")
					}
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
		qstring := regexp.MustCompile(`(?s)^"(.*)"$`)
		integer := regexp.MustCompile(`^[+-]?\d+(\.\d+)?$`)
		float := regexp.MustCompile(`^[+-]?\d*\.\d+$`)
		envvar := regexp.MustCompile(`^\$[a-zA-Z_][a-zA-Z0-9_.]*$`)

		var final []*Expr
		var left, op *Expr

		pop := func() {
			if left != nil {
				final = append(final, left)
				left = nil
			}
		}

		push := func(e *Expr) {
			TRACE("expr: pushing data expression `%s' onto stack", e)
			TRACE("expr:   start: left=`%s', op=`%s'", left, op)
			defer func() { TRACE("expr:     end: left=`%s', op=`%s'\n", left, op) }()

			if left == nil {
				left = e
				return
			}
			if op == nil {
				pop()
				left = e
				return
			}
			op.Left = left
			op.Right = e
			left = op
			op = nil
		}

		TRACE("expr: parsing `%s'", src)
		for i, arg := range split(src) {
			switch {
			case arg == ",":
				DEBUG("  #%d: literal comma found; treating what we've seen so far as a complete expression", i)
				pop()

			case envvar.MatchString(arg):
				DEBUG("  #%d: parsed as unquoted environment variable reference '%s'", i, arg)
				push(&Expr{Type: EnvVar, Name: arg[1:]})

			case qstring.MatchString(arg):
				m := qstring.FindStringSubmatch(arg)
				DEBUG("  #%d: parsed as quoted string literal '%s'", i, m[1])
				push(&Expr{Type: Literal, Literal: m[1]})

			case float.MatchString(arg):
				DEBUG("  #%d: parsed as unquoted floating point literal '%s'", i, arg)
				v, err := strconv.ParseFloat(arg, 64)
				if err != nil {
					DEBUG("  #%d: %s is not parsable as a floating point number: %s", i, arg, err)
					return args, err
				}
				push(&Expr{Type: Literal, Literal: v})

			case integer.MatchString(arg):
				DEBUG("  #%d: parsed as unquoted integer literal '%s'", i, arg)
				v, err := strconv.ParseInt(arg, 10, 64)
				if err == nil {
					push(&Expr{Type: Literal, Literal: v})
					break
				}
				DEBUG("  #%d: %s is not parsable as an integer, falling back to parsing as float: %s", i, arg, err)
				f, err := strconv.ParseFloat(arg, 64)
				push(&Expr{Type: Literal, Literal: f})
				if err != nil {
					panic("Could not actually parse as an int or a float. Need to fix regexp?")
				}

			case arg == "||":
				DEBUG("  #%d: parsed logical-or operator, '||'", i)

				if left == nil || op != nil {
					return args, fmt.Errorf(`syntax error near: %s`, src)
				}
				TRACE("expr: pushing || expr-op onto the stack")
				op = &Expr{Type: LogicalOr}

			case arg == "nil" || arg == "null" || arg == "~" || arg == "Nil" || arg == "Null" || arg == "NIL" || arg == "NULL":
				DEBUG("  #%d: parsed the nil value token '%s'", i, arg)
				push(&Expr{Type: Literal, Literal: nil})

			case arg == "false" || arg == "False" || arg == "FALSE":
				DEBUG("  #%d: parsed the false value token '%s'", i, arg)
				push(&Expr{Type: Literal, Literal: false})

			case arg == "true" || arg == "True" || arg == "TRUE":
				DEBUG("  #%d: parsed the true value token '%s'", i, arg)
				push(&Expr{Type: Literal, Literal: true})

			default:
				c, err := tree.ParseCursor(arg)
				if err != nil {
					DEBUG("  #%d: %s is a malformed reference: %s", i, arg, err)
					return args, err
				}
				DEBUG("  #%d: parsed as a reference to $.%s", i, c)
				push(&Expr{Type: Reference, Reference: c, BracketedNodes: bracketsOf(arg)})
			}
		}
		pop()
		if left != nil || op != nil {
			return nil, fmt.Errorf(`syntax error near: %s`, src)
		}
		DEBUG("")

		for _, e := range final {
			TRACE("expr: pushing expression `%v' onto the operand list", e)
			reduced, err := e.Reduce()
			if err != nil {
				if warning, isWarning := err.(WarningError); isWarning {
					warning.Warn()
				} else {
					fmt.Fprintf(os.Stdout, "warning: %s\n", err)
				}
			}
			args = append(args, reduced)
		}

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
		if _, ok := op.op.(NullOperator); ok && len(m[2]) == 0 {
			DEBUG("skipping `%s': not a real operator -- might be a BOSH variable?", src)
			continue
		}
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
func (op *Opcall) Dependencies(ev *Evaluator, locs []*tree.Cursor) []*tree.Cursor {
	l := []*tree.Cursor{}
	for _, arg := range op.args {
		l = append(l, arg.Dependencies(ev, locs)...)
	}

	return op.op.Dependencies(ev, op.args, locs, l)
}

// Run ...
func (op *Opcall) Run(ev *Evaluator) (*Response, error) {
	was := ev.Here
	ev.Here = op.where
	r, err := op.op.Run(ev, op.args)
	ev.Here = was

	if err != nil {
		return nil, ansi.Errorf("@m{$.%s}: @R{%s}", op.where, err)
	}
	return r, nil
}
