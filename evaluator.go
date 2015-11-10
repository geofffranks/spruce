package main

import (
	"fmt"
	"strconv"
)

// Evaluator ...
type Evaluator struct {
	Tree map[interface{}]interface{}
	Deps map[string][]Cursor

	Here *Cursor

	DataOps  []*Opcall
	CheckOps []*Opcall

	pointer *interface{}
}

// DataFlow ...
func (ev *Evaluator) DataFlow() error {
	ev.Here = &Cursor{}

	all := map[string]*Opcall{}
	locs := []*Cursor{}
	errors := MultiError{Errors: []error{}}

	// forward decls of co-recursive function
	var check func(interface{})
	var scan func(interface{})

	check = func(v interface{}) {
		if s, ok := v.(string); ok {
			op, err := ParseOpcall(s)
			if err != nil {
				errors.Append(err)
			} else if op != nil {
				if _, ok := op.op.(ParamOperator); !ok {
					op.where = ev.Here.Copy()
					if canon, err := op.where.Canonical(ev.Tree); err == nil {
						op.canonical = canon
					} else {
						op.canonical = op.where
					}
					all[op.canonical.String()] = op
					locs = append(locs, op.canonical)
				}
			}
		} else {
			scan(v)
		}
	}

	scan = func(o interface{}) {
		switch o.(type) {
		case map[interface{}]interface{}:
			for k, v := range o.(map[interface{}]interface{}) {
				ev.Here.Push(k.(string))
				check(v)
				ev.Here.Pop()
			}

		case []interface{}:
			for i, v := range o.([]interface{}) {
				ev.Here.Push(fmt.Sprintf("%d", i))
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

	// find all nodes in g that are free (no futher dependencies)
	freeNodes := func(g [][2]*Opcall) []*Opcall {
		l := []*Opcall{}
		for k, node := range all {
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
	ev.DataOps = []*Opcall{} // order in which to call the ops
	for len(all) > 0 {
		free := freeNodes(g)
		if len(free) == 0 {
			return fmt.Errorf("cycle detected in operator data-flow graph")
		}

		for _, node := range free {
			ev.DataOps = append(ev.DataOps, node)
			g = remove(g, node)
		}
	}

	if len(errors.Errors) > 0 {
		return errors
	}
	return nil
}

// Patch ...
func (ev *Evaluator) Patch() error {
	DEBUG("patching up YAML by evaluating outstanding operators\n")

	errors := MultiError{Errors: []error{}}
	for _, op := range ev.DataOps {
		resp, err := op.Run(ev)
		if err != nil {
			errors.Append(err)
			continue
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
				errors.Append(err)
				continue
			}
			switch o.(type) {
			case []interface{}:
				i, err := strconv.ParseUint(key, 10, 0)
				if err != nil {
					// FIXME: handle named?
					DEBUG("  error: %s\n  continuing\n", err)
					errors.Append(err)
					continue
				}
				o.([]interface{})[i] = resp.Value

			case map[interface{}]interface{}:
				o.(map[interface{}]interface{})[key] = resp.Value

			default:
				err := TypeMismatchError{
					Path:   parent.Nodes,
					Wanted: "a map or a list",
					Got:    "a scalar",
				}
				DEBUG("  error: %s\n  continuing\n", err)
				errors.Append(err)
				continue
			}
			DEBUG("")

		case Inject:
			DEBUG("executing an Inject instruction on %s", op.where)
			key := op.where.Component(-1)
			parent := op.where.Copy()
			parent.Pop()

			o, err := parent.Resolve(ev.Tree)
			if err != nil {
				errors.Append(err)
				DEBUG("  error: %s\n  continuing\n", err)
				continue
			}

			m := o.(map[interface{}]interface{})
			delete(m, key)

			for k, v := range resp.Value.(map[interface{}]interface{}) {
				_, set := m[k]
				if !set {
					DEBUG("  %s is not set, using the injected value", parent)
					m[k] = v
				} else {
					DEBUG("  %s exists, merging the injected values", parent)
					target, targetIsMap := m[k].(map[interface{}]interface{})
					template, templateIsMap := v.(map[interface{}]interface{})
					if targetIsMap && templateIsMap {
						m[k], err = Merge(template, target)
						if err != nil {
							DEBUG("  error: %s\n  continuing\n", err)
							errors.Append(err)
						}
					}
				}
			}
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
		c, err := ParseCursor(path)
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

		// FIXME: should we handle array index removal?

		default:
			DEBUG("  I don't know how to prune %s\n    value=%v\n", path, o)
		}
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
			return fmt.Errorf("Hit max recursion depth. You seem to have a self-referencing dataset")
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

// CheckParams ...
func (ev *Evaluator) CheckParams() error {
	DEBUG("checking for any remaining (( param ... )) operators\n")

	ev.Here = &Cursor{}
	ev.CheckOps = []*Opcall{}

	errors := MultiError{Errors: []error{}}

	var check func(interface{})
	var scan func(interface{})

	check = func(v interface{}) {
		if s, ok := v.(string); ok {
			op, err := ParseOpcall(s)
			if err != nil {
				errors.Append(err)

			} else if op != nil {
				if _, ok := op.op.(ParamOperator); ok {
					op.where = ev.Here.Copy()
					ev.CheckOps = append(ev.CheckOps, op)
				}
			}
		} else {
			scan(v)
		}
	}

	scan = func(o interface{}) {
		switch o.(type) {
		case map[interface{}]interface{}:
			for k, v := range o.(map[interface{}]interface{}) {
				ev.Here.Push(k.(string))
				check(v)
				ev.Here.Pop()
			}

		case []interface{}:
			for i, v := range o.([]interface{}) {
				ev.Here.Push(fmt.Sprintf("%d", i))
				check(v)
				ev.Here.Pop()
			}
		}

		return
	}

	scan(ev.Tree)
	for _, op := range ev.CheckOps {
		_, err := op.Run(ev)
		if err != nil {
			errors.Append(err)
		}
	}
	if len(errors.Errors) == 0 {
		return nil
	}
	return errors
}

// Run ...
func (ev *Evaluator) Run(prune []string) error {
	errors := MultiError{Errors: []error{}}
	errors.Append(ev.DataFlow())
	errors.Append(ev.Patch())
	errors.Append(ev.Prune(prune))

	// this is a big failure...
	if err := ev.CheckForCycles(4096); err != nil {
		return err
	}

	errors.Append(ev.CheckParams())
	if len(errors.Errors) > 0 {
		return errors
	}
	return nil
}
