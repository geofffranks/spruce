package spruce

import (
	"fmt"
	"github.com/starkandwayne/goutils/ansi"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/starkandwayne/goutils/tree"

	. "github.com/geofffranks/spruce/log"
)

// FileOperator ...
type FileOperator struct{}

// Setup ...
func (FileOperator) Setup() error {
	return nil
}

// Phase ...
func (FileOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (FileOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor) []*tree.Cursor {
	return []*tree.Cursor{}
}

// Run ...
func (FileOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( file ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( file ... )) operation at $%s\n", ev.Here)

	if len(args) != 1 {
		return nil, fmt.Errorf("file operator requires exactly one string or reference argument")
	}

	var fname string
	fbasepath := os.Getenv("SPRUCE_FILE_BASE_PATH")

	arg := args[0]
	i := 0
	v, err := arg.Resolve(ev.Tree)
	if err != nil {
		DEBUG("  arg[%d]: failed to resolve expression to a concrete value", i)
		DEBUG("     [%d]: error was: %s", i, err)
		return nil, err
	}

	switch v.Type {
	case Literal:
		DEBUG("  arg[%d]: using string literal '%v'", i, v.Literal)
		DEBUG("     [%d]: appending '%v' to resultant string", i, v.Literal)
		fname = fmt.Sprintf("%v", v.Literal)

	case Reference:
		DEBUG("  arg[%d]: trying to resolve reference $.%s", i, v.Reference)
		s, err := v.Reference.Resolve(ev.Tree)
		if err != nil {
			DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
			return nil, fmt.Errorf("Unable to resolve `%s`: %s", v.Reference, err)
		}

		switch s.(type) {
		case map[interface{}]interface{}:
			DEBUG("  arg[%d]: %v is not a string scalar", i, s)
			return nil, ansi.Errorf("@R{tried to read file} @c{%s}@R{, which is not a string scalar}", v.Reference)

		case []interface{}:
			DEBUG("  arg[%d]: %v is not a string scalar", i, s)
			return nil, ansi.Errorf("@R{tried to read file} @c{%s}@R{, which is not a string scalar}", v.Reference)

		default:
			DEBUG("     [%d]: appending '%s' to resultant string", i, s)
			fname = fmt.Sprintf("%v", s)
		}

	default:
		DEBUG("  arg[%d]: I don't know what to do with '%v'", i, arg)
		return nil, fmt.Errorf("file operator only accepts string literals and key reference argument")
	}
	DEBUG("")

	if !filepath.IsAbs(fname) {
		fname = filepath.Join(fbasepath, fname)
	}

	contents, err := ioutil.ReadFile(fname)
	if err != nil {
		DEBUG("  File %s cannot be read: %s", fname, err)
		return nil, ansi.Errorf("@R{tried to read file} @c{%s}@R{: could not be read - %s}", fname, err)
	}

	DEBUG("  resolved (( file ... )) operation to the string:\n    \"%s\"", string(contents))

	return &Response{
		Type:  Replace,
		Value: string(contents),
	}, nil
}

func init() {
	RegisterOp("file", FileOperator{})
}
