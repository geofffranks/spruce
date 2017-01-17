package spruce

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/starkandwayne/goutils/ansi"
	"github.com/starkandwayne/goutils/tree"

	. "github.com/geofffranks/spruce/log"
)

// CalcOperator is invoked with (( calc <expression> ))
type CalcOperator struct{}

// Setup ...
func (CalcOperator) Setup() error {
	return nil
}

// Phase ...
func (CalcOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies ...
func (CalcOperator) Dependencies(ev *Evaluator, args []*Expr, _ []*tree.Cursor) []*tree.Cursor {
	DEBUG("Calculating dependencies for (( calc ... ))")
	deps := []*tree.Cursor{}

	// The dependencies are checks straigth forward on the happy path:
	// There must be one literal argument containing possible references.
	if len(args) == 1 {
		switch args[0].Type {
		case Literal:
			if cursors, searchError := searchForCursors(args[0].Literal.(string)); searchError == nil {
				for _, cursor := range cursors {
					deps = append(deps, cursor)
				}
			}
		}
	}

	return deps
}

// Run ...
func (CalcOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	DEBUG("running (( calc ... )) operation at $.%s", ev.Here)
	defer DEBUG("done with (( calc ... )) operation at $%s\n", ev.Here)

	// The calc operator expects one literal argument containing the quoted expression to be evaluated
	if len(args) != 1 {
		return nil, ansi.Errorf("@R{calc operator only expects} @r{one} @R{argument containing the expression}")
	}

	switch args[0].Type {
	case Literal:
		// Replace all Spruce references with the respective value
		DEBUG("  input expression: %s", args[0].Literal.(string))
		input, replaceError := replaceReferences(ev, args[0].Literal.(string))
		if replaceError != nil {
			return nil, replaceError
		}

		// Once all Spruce references (variables) are replaced, try to read the expression
		DEBUG("  processed expression: %s", input)
		expression, expressionError := govaluate.NewEvaluableExpressionWithFunctions(input, supportedFunctions())
		if expressionError != nil {
			return nil, expressionError
		}

		// Check that there are no named variables in the expression that we cannot evaluate/insert
		if len(expression.Vars()) > 0 {
			return nil, ansi.Errorf("@R{calc operator does not support named variables in expression:} @r{%s}", strings.Join(expression.Vars(), ", "))
		}

		// Evaluate without a variables list (named variables are not supported)
		result, evaluateError := expression.Evaluate(nil)
		if evaluateError != nil {
			return nil, evaluateError
		}

		DEBUG("  evaluated result: %v", result)
		return &Response{
			Type:  Replace,
			Value: result,
		}, nil

	default:
		return nil, ansi.Errorf("@R{calc operator argument is suppose to be a quoted mathematical expression (type} @r{Literal}@R{)}")
	}
}

func searchForCursors(input string) ([]*tree.Cursor, error) {
	result := []*tree.Cursor{}

	// Search for sub-strings that contain the path seperator dot character
	// https://regex101.com/r/TIEyak/1 (to delete the URL use https://regex101.com/delete/fPbxosYXWzBPYaNdL5YcPpj3)
	regexp := regexp.MustCompile(`(\w+|-)\.(\w+|-|\.)+`)
	candidates := regexp.FindAllString(input, -1)
	DEBUG("    strings found containing the path seperator: %v", strings.Join(candidates, ", "))

	// If it is a path, it can be parsed (parse errors will be ignored)
	for _, candidate := range candidates {
		// Skip floats
		if _, err := strconv.ParseFloat(candidate, 64); err == nil {
			continue
		}

		// Skip ints
		if _, err := strconv.ParseInt(candidate, 10, 64); err == nil {
			continue
		}

		if cursor, parseError := tree.ParseCursor(candidate); parseError == nil {
			result = append(result, cursor)
		}
	}

	DEBUG("    result cursors: %v", result)
	return result, nil
}

func replaceReferences(ev *Evaluator, input string) (string, error) {
	cursors, searchError := searchForCursors(input)
	if searchError != nil {
		return "", searchError
	}

	for _, cursor := range cursors {
		value, resolveError := cursor.Resolve(ev.Tree)
		if resolveError != nil {
			return "", resolveError
		}

		path := cursor.String()
		DEBUG("    path/value: %s=%v", path, value)

		switch value.(type) {
		case int, uint8, uint16, uint32, uint64, int8, int16, int32, int64:
			input = strings.Replace(input, path, fmt.Sprintf("%d", value), -1)

		case float32, float64:
			input = strings.Replace(input, path, fmt.Sprintf("%f", value), -1)

		case nil:
			return "", ansi.Errorf("@R{path} @r{%s} @R{references a }@r{nil}@R{ value, which cannot be used in calculations}", path)

		default:
			return "", ansi.Errorf("@R{path} @r{%s} @R{is of type} @r{%s}@R{, which cannot be used in calculations}", path, reflect.TypeOf(value).Kind())
		}
	}

	return input, nil
}

func supportedFunctions() map[string]govaluate.ExpressionFunction {
	return map[string]govaluate.ExpressionFunction{
		"min": func(args ...interface{}) (interface{}, error) {
			if len(args) == 2 && reflect.TypeOf(args[0]).Kind() == reflect.Float64 && reflect.TypeOf(args[1]).Kind() == reflect.Float64 {
				return math.Min(args[0].(float64), args[1].(float64)), nil

			} else {
				return -1, ansi.Errorf("@R{min function expects} @r{two arguments} @R{of type} @r{float64}")
			}
		},

		"max": func(args ...interface{}) (interface{}, error) {
			if len(args) == 2 && reflect.TypeOf(args[0]).Kind() == reflect.Float64 && reflect.TypeOf(args[1]).Kind() == reflect.Float64 {
				return math.Max(args[0].(float64), args[1].(float64)), nil

			} else {
				return -1, ansi.Errorf("@R{max function expects} @r{two arguments} @R{of type} @r{float64}")
			}
		},

		"mod": func(args ...interface{}) (interface{}, error) {
			if len(args) == 2 && reflect.TypeOf(args[0]).Kind() == reflect.Float64 && reflect.TypeOf(args[1]).Kind() == reflect.Float64 {
				return math.Mod(args[0].(float64), args[1].(float64)), nil

			} else {
				return -1, ansi.Errorf("@R{mod function expects} @r{two arguments} @R{of type} @r{float64}")
			}
		},

		"pow": func(args ...interface{}) (interface{}, error) {
			if len(args) == 2 && reflect.TypeOf(args[0]).Kind() == reflect.Float64 && reflect.TypeOf(args[1]).Kind() == reflect.Float64 {
				return math.Pow(args[0].(float64), args[1].(float64)), nil

			} else {
				return -1, ansi.Errorf("@R{pow function expects} @r{two arguments} @R{of type} @r{float64}")
			}
		},

		"sqrt": func(args ...interface{}) (interface{}, error) {
			if len(args) == 1 && reflect.TypeOf(args[0]).Kind() == reflect.Float64 {
				return math.Sqrt(args[0].(float64)), nil

			} else {
				return -1, ansi.Errorf("@R{sqrt function expects} @r{one argument} @R{of type} @r{float64}")
			}
		},

		"floor": func(args ...interface{}) (interface{}, error) {
			if len(args) == 1 && reflect.TypeOf(args[0]).Kind() == reflect.Float64 {
				return math.Floor(args[0].(float64)), nil

			} else {
				return -1, ansi.Errorf("@R{floor function expects} @r{one argument} @R{of type} @r{float64}")
			}
		},

		"ceil": func(args ...interface{}) (interface{}, error) {
			if len(args) == 1 && reflect.TypeOf(args[0]).Kind() == reflect.Float64 {
				return math.Ceil(args[0].(float64)), nil

			} else {
				return -1, ansi.Errorf("@R{ceil function expects} @r{one argument} @R{of type} @r{float64}")
			}
		},
	}
}

func init() {
	RegisterOp("calc", CalcOperator{})
}
