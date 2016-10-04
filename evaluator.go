package spruce

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/starkandwayne/goutils/ansi"

	. "github.com/geofffranks/spruce/log"
	"github.com/starkandwayne/goutils/tree"
)

// Evaluator ...
type Evaluator struct {
	Tree map[interface{}]interface{}
	Deps map[string][]tree.Cursor

	Here *tree.Cursor

	CheckOps []*Opcall

	pointer *interface{}
}

func nameOfObj(o interface{}, def string) string {
	for _, field := range tree.NameFields {
		switch o.(type) {
		case map[string]interface{}:
			if value, ok := o.(map[string]interface{})[field]; ok {
				if s, ok := value.(string); ok {
					return s
				}
			}
		case map[interface{}]interface{}:
			if value, ok := o.(map[interface{}]interface{})[field]; ok {
				if s, ok := value.(string); ok {
					return s
				}
			}
		}
	}
	return def
}

// DataFlow ...
func (ev *Evaluator) DataFlow(phase OperatorPhase) ([]*Opcall, error) {
	ev.Here = &tree.Cursor{}

	all := map[string]*Opcall{}
	locs := []*tree.Cursor{}
	errors := MultiError{Errors: []error{}}

	// forward decls of co-recursive function
	var check func(interface{})
	var scan func(interface{})

	check = func(v interface{}) {
		if s, ok := v.(string); ok {
			op, err := ParseOpcall(phase, s)
			if err != nil {
				errors.Append(err)
			} else if op != nil {
				op.where = ev.Here.Copy()
				if canon, err := op.where.Canonical(ev.Tree); err == nil {
					op.canonical = canon
				} else {
					op.canonical = op.where
				}
				all[op.canonical.String()] = op
				TRACE("found an operation at %s: %s", op.where.String(), op.src)
				TRACE("        (canonical at %s)", op.canonical.String())
				locs = append(locs, op.canonical)
			}
		} else {
			scan(v)
		}
	}

	scan = func(o interface{}) {
		switch o.(type) {
		case map[interface{}]interface{}:
			for k, v := range o.(map[interface{}]interface{}) {
				ev.Here.Push(fmt.Sprintf("%v", k))
				check(v)
				ev.Here.Pop()
			}

		case []interface{}:
			for i, v := range o.([]interface{}) {
				name := nameOfObj(v, fmt.Sprintf("%d", i))
				op, _ := ParseOpcall(phase, name)
				if op == nil {
					ev.Here.Push(name)
				} else {
					ev.Here.Push(fmt.Sprintf("%d", i))
				}
				check(v)
				ev.Here.Pop()
			}
		}
	}

	scan(ev.Tree)

	// construct the data flow graph, where a -> b = b calls/requires a
	// represent the graph as list of adjancies, that is [a,b] = a -> b
	var g [][2]*Opcall
	for _, a := range all {
		for _, path := range a.Dependencies(ev, locs) {
			if b, found := all[path.String()]; found {
				g = append(g, [2]*Opcall{b, a})
			}
		}
	}

	// construct a sorted list of keys in $all, so that we
	// can reliably generate the same DFA every time
	var sortedKeys []string
	for k := range all {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	// find all nodes in g that are free (no futher dependencies)
	freeNodes := func(g [][2]*Opcall) []*Opcall {
		l := []*Opcall{}
		for _, k := range sortedKeys {
			node, ok := all[k]
			if !ok {
				continue
			}

			called := false
			for _, pair := range g {
				if pair[1] == node {
					called = true
					break
				}
			}

			if !called {
				delete(all, k)
				l = append(l, node)
			}
		}

		return l
	}

	// removes (nullifies) all dependencies on n in g
	remove := func(old [][2]*Opcall, n *Opcall) [][2]*Opcall {
		l := [][2]*Opcall{}
		for _, pair := range old {
			if pair[0] != n {
				l = append(l, pair)
			}
		}
		return l
	}

	// Kahn topological sort
	ops := []*Opcall{} // order in which to call the ops
	wave := 0
	for len(all) > 0 {
		wave++
		free := freeNodes(g)
		if len(free) == 0 {
			return nil, ansi.Errorf("@*{cycle detected in operator data-flow graph}")
		}

		for _, node := range free {
			TRACE("data flow: [%d] wave %d, op %s: %s", len(ops), wave, node.where, node.src)
			ops = append(ops, node)
			g = remove(g, node)
		}
	}

	if len(errors.Errors) > 0 {
		return nil, errors
	}
	return ops, nil
}

// RunOps ...
func (ev *Evaluator) RunOps(ops []*Opcall) error {
	DEBUG("patching up YAML by evaluating outstanding operators\n")

	errors := MultiError{Errors: []error{}}
	for _, op := range ops {
		err := ev.RunOp(op)
		if err != nil {
			errors.Append(err)
		}
	}

	if len(errors.Errors) > 0 {
		return errors
	}
	return nil
}

// Prune ...
func (ev *Evaluator) Prune(paths []string) error {
	DEBUG("pruning %d paths from the final YAML structure", len(paths))
	for _, path := range paths {
		c, err := tree.ParseCursor(path)
		if err != nil {
			return err
		}

		key := c.Component(-1)
		parent := c.Copy()
		parent.Pop()
		o, err := parent.Resolve(ev.Tree)
		if err != nil {
			continue
		}

		switch o.(type) {
		case map[interface{}]interface{}:
			if _, ok := o.(map[interface{}]interface{}); ok {
				DEBUG("  pruning %s", path)
				delete(o.(map[interface{}]interface{}), key)
			}

		case []interface{}:
			if list, ok := o.([]interface{}); ok {
				if idx, err := strconv.Atoi(key); err == nil {
					parent.Pop()
					if s, err := parent.Resolve(ev.Tree); err == nil {
						if reflect.TypeOf(s).Kind() == reflect.Map {
							parentName := fmt.Sprintf("%s", c.Component(-2))
							DEBUG("  pruning index %d of array '%s'", idx, parentName)

							length := len(list) - 1
							replacement := make([]interface{}, length)
							copy(replacement, append(list[:idx], list[idx+1:]...))

							delete(s.(map[interface{}]interface{}), parentName)
							s.(map[interface{}]interface{})[parentName] = replacement
						}
					}
				}
			}

		default:
			DEBUG("  I don't know how to prune %s\n    value=%v\n", path, o)
		}
	}
	DEBUG("")
	return nil
}

// Cherry-pick ...
func (ev *Evaluator) CherryPick(paths []string) error {
	DEBUG("cherry-picking %d paths from the final YAML structure", len(paths))

  if len(paths) > 0 {
		// This will serve as the replacement tree ...
		replacement := make(map[interface{}]interface{})

		for _, path := range paths {
			cursor, err := tree.ParseCursor(path)
			if err != nil {
				return err
			}

			// These variables will potentially be modified (depending on the structure)
			var cherryName string
			var cherryValue interface{}

			// Resolve the value that needs to be cherry picked
			cherryValue, err = cursor.Resolve(ev.Tree)
			if err != nil {
				return err
			}

			// Name of the parameter of the to-be-picked value
			cherryName = cursor.Nodes[len(cursor.Nodes)-1]

			// Since the cherry can be deep down the structure, we need to go down
			// (or up, depending how you read it) the structure to include the parent
			// names of the respective cherry. The pointer will be reassigned with
			// each level.
			pointer := cursor
			for pointer != nil {
				parent := pointer.Copy()
				parent.Pop()

				if parent.String() == "" {
					// Empty parent string means we reached the root, setting the pointer nil to stop processing ...
					pointer = nil

					// ... and adding the cherry to the replacement map
					DEBUG("Adding '%s' to the replacement tree", path)
					replacement[cherryName] = cherryValue

				} else {
					// Reassign the pointer to the parent and restructre the current cherry value to address the parent structure and name
					pointer = parent

					// Depending on the type of the parent, either a map or a list is created for the new parent of the cherry value
					if obj, err := parent.Resolve(ev.Tree); err == nil {
						switch obj.(type) {
						case map[interface{}]interface{}:
							tmp := make(map[interface{}]interface{})
							tmp[cherryName] = cherryValue

							cherryName = parent.Nodes[len(parent.Nodes)-1]
							cherryValue = tmp

						case []interface{}:
							tmp := make([]interface{}, 0, 0)
							tmp = append(tmp, cherryValue)

							cherryName = parent.Nodes[len(parent.Nodes)-1]
							cherryValue = tmp

						default:
							return ansi.Errorf("@*{Unsupported type detected, %s is neither a map nor a list}", parent.String())
						}

					} else {
						return err
					}
				}
			}
		}

		// replace the existing tree with a new one that contain the cherry-picks
		ev.Tree = replacement
	}

	DEBUG("")
	return nil
}

// CheckForCycles ...
func (ev *Evaluator) CheckForCycles(maxDepth int) error {
	DEBUG("checking for cycles in final YAML structure")

	var check func(o interface{}, depth int) error
	check = func(o interface{}, depth int) error {
		if depth == 0 {
			return ansi.Errorf("@*{Hit max recursion depth. You seem to have a self-referencing dataset}")
		}

		switch o.(type) {
		case []interface{}:
			for _, v := range o.([]interface{}) {
				if err := check(v, depth-1); err != nil {
					return err
				}
			}

		case map[interface{}]interface{}:
			for _, v := range o.(map[interface{}]interface{}) {
				if err := check(v, depth-1); err != nil {
					return err
				}
			}
		}

		return nil
	}

	err := check(ev.Tree, maxDepth)
	if err != nil {
		DEBUG("error: %s\n", err)
		return err
	}

	DEBUG("no cycles detected.\n")
	return nil
}

// RunOp ...
func (ev *Evaluator) RunOp(op *Opcall) error {
	resp, err := op.Run(ev)
	if err != nil {
		return err
	}

	switch resp.Type {
	case Replace:
		DEBUG("executing a Replace instruction on %s", op.where)
		key := op.where.Component(-1)
		parent := op.where.Copy()
		parent.Pop()

		o, err := parent.Resolve(ev.Tree)
		if err != nil {
			DEBUG("  error: %s\n  continuing\n", err)
			return err
		}
		switch o.(type) {
		case []interface{}:
			i, err := strconv.ParseUint(key, 10, 0)
			if err != nil {
				DEBUG("  error: %s\n  continuing\n", err)
				return err
			}
			o.([]interface{})[i] = resp.Value

		case map[interface{}]interface{}:
			o.(map[interface{}]interface{})[key] = resp.Value

		default:
			err := tree.TypeMismatchError{
				Path:   parent.Nodes,
				Wanted: "a map or a list",
				Got:    "a scalar",
			}
			DEBUG("  error: %s\n  continuing\n", err)
			return err
		}
		DEBUG("")

	case Inject:
		DEBUG("executing an Inject instruction on %s", op.where)
		key := op.where.Component(-1)
		parent := op.where.Copy()
		parent.Pop()

		o, err := parent.Resolve(ev.Tree)
		if err != nil {
			DEBUG("  error: %s\n  continuing\n", err)
			return err
		}

		m := o.(map[interface{}]interface{})
		delete(m, key)

		for k, v := range resp.Value.(map[interface{}]interface{}) {
			path := fmt.Sprintf("%s.%s", parent, k)
			_, set := m[k]
			if !set {
				DEBUG("  %s is not set, using the injected value", path)
				m[k] = v
			} else {
				DEBUG("  %s is set, merging the injected value", path)
				merger := &Merger{AppendByDefault: true}
				merged := merger.mergeObj(v, m[k], path)
				if err := merger.Error(); err != nil {
					return err
				}
				m[k] = merged
			}
		}
	}
	return nil
}

// RunPhase ...
func (ev *Evaluator) RunPhase(p OperatorPhase) error {
	err := SetupOperators(p)
	if err != nil {
		return err
	}
	op, err := ev.DataFlow(p)
	if err != nil {
		return err
	}
	return ev.RunOps(op)
}

// Run ...
func (ev *Evaluator) Run(prune []string, picks []string) error {
	errors := MultiError{Errors: []error{}}
	paramErrs := MultiError{Errors: []error{}}

	errors.Append(ev.RunPhase(MergePhase))
	paramErrs.Append(ev.RunPhase(ParamPhase))
	if len(paramErrs.Errors) > 0 {
		return paramErrs
	}

	errors.Append(ev.RunPhase(EvalPhase))

	// this is a big failure...
	if err := ev.CheckForCycles(4096); err != nil {
		return err
	}

	// post-processing: prune
	addToPruneListIfNecessary(prune...)
	errors.Append(ev.Prune(keysToPrune))
	keysToPrune = nil

	// post-processing: cherry-pick
	errors.Append(ev.CherryPick(picks))

	if len(errors.Errors) > 0 {
		return errors
	}
	return nil
}
