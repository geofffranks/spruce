package tree

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Cursor ...
type Cursor struct {
	Nodes []string
}

func listFind(l []interface{}, fields []string, key string) (interface{}, uint64, bool) {
	for _, field := range fields {
		for i, v := range l {
			switch v.(type) {
			case map[string]interface{}:
				value, ok := v.(map[string]interface{})[field]
				if ok && value == key {
					return v, uint64(i), true
				}
			case map[interface{}]interface{}:
				value, ok := v.(map[interface{}]interface{})[field]
				if ok && value == key {
					return v, uint64(i), true
				}
			}
		}
	}
	return nil, 0, false
}

// ParseCursor ...
func ParseCursor(s string) (*Cursor, error) {
	var nodes []string
	node := bytes.NewBuffer([]byte{})
	bracketed := false

	push := func() {
		if node.Len() == 0 {
			return
		}
		s := node.String()
		if len(nodes) == 0 && s == "$" {
			node.Reset()
			return
		}

		nodes = append(nodes, s)
		node.Reset()
	}

	for pos, c := range s {
		switch c {
		case '.':
			if bracketed {
				node.WriteRune(c)
			} else {
				push()
			}

		case '[':
			if bracketed {
				return nil, SyntaxError{
					Problem:  "unexpected '['",
					Position: pos,
				}
			}
			push()
			bracketed = true

		case ']':
			if !bracketed {
				return nil, SyntaxError{
					Problem:  "unexpected ']'",
					Position: pos,
				}
			}
			push()
			bracketed = false

		default:
			node.WriteRune(c)
		}
	}
	push()

	return &Cursor{
		Nodes: nodes,
	}, nil
}

// Copy ...
func (c *Cursor) Copy() *Cursor {
	other := &Cursor{Nodes: []string{}}
	for _, node := range c.Nodes {
		other.Nodes = append(other.Nodes, node)
	}
	return other
}

// Under ...
func (c *Cursor) Under(other *Cursor) bool {
	if len(c.Nodes) <= len(other.Nodes) {
		return false
	}
	match := false
	for i := range other.Nodes {
		if c.Nodes[i] != other.Nodes[i] {
			return false
		}
		match = true
	}
	return match
}

// Pop ...
func (c *Cursor) Pop() string {
	if len(c.Nodes) == 0 {
		return ""
	}
	last := c.Nodes[len(c.Nodes)-1]
	c.Nodes = c.Nodes[0 : len(c.Nodes)-1]
	return last
}

// Push ...
func (c *Cursor) Push(n string) {
	c.Nodes = append(c.Nodes, n)
}

// String ...
func (c *Cursor) String() string {
	return strings.Join(c.Nodes, ".")
}

// Depth ...
func (c *Cursor) Depth() int {
	return len(c.Nodes)
}

// Parent ...
func (c *Cursor) Parent() string {
	if len(c.Nodes) < 2 {
		return ""
	}
	return c.Nodes[len(c.Nodes)-2]
}

// Component ...
func (c *Cursor) Component(offset int) string {
	offset = len(c.Nodes) + offset
	if offset < 0 || offset >= len(c.Nodes) {
		return ""
	}
	return c.Nodes[offset]
}

// Canonical ...
func (c *Cursor) Canonical(o interface{}) (*Cursor, error) {
	canon := &Cursor{Nodes: []string{}}

	for _, k := range c.Nodes {
		switch o.(type) {
		case []interface{}:
			i, err := strconv.ParseUint(k, 10, 0)
			if err == nil {
				// if k is an integer (in string form), go by index
				if int(i) >= len(o.([]interface{})) {
					return nil, NotFoundError{
						Path: canon.Nodes,
					}
				}
				o = o.([]interface{})[i]

			} else {
				// if k is a string, look for immediate map descendants who have
				//     'name', 'key' or 'id' fields matching k
				var found bool
				o, i, found = listFind(o.([]interface{}), []string{"name", "key", "id"}, k)
				if !found {
					return nil, NotFoundError{
						Path: canon.Nodes,
					}
				}
			}
			canon.Push(fmt.Sprintf("%d", i))

		case map[string]interface{}:
			canon.Push(k)
			var ok bool
			o, ok = o.(map[string]interface{})[k]
			if !ok {
				return nil, NotFoundError{
					Path: canon.Nodes,
				}
			}

		case map[interface{}]interface{}:
			canon.Push(k)
			var ok bool
			o, ok = o.(map[interface{}]interface{})[k]
			if !ok {
				return nil, NotFoundError{
					Path: canon.Nodes,
				}
			}

		default:
			return nil, TypeMismatchError{
				Path:   canon.Nodes,
				Wanted: "a map or a list",
				Got:    "a scalar",
			}
		}
	}

	return canon, nil
}

// Glob ...
func (c *Cursor) Glob(tree interface{}) ([]*Cursor, error) {
	var resolver func(interface{}, []string, []string, int) ([]interface{}, error)
	resolver = func(o interface{}, here, path []string, pos int) ([]interface{}, error) {
		if pos == len(path) {
			return []interface{}{
				(&Cursor{Nodes: here}).Copy(),
			}, nil
		}

		paths := []interface{}{}
		k := path[pos]
		if k == "*" {
			switch o.(type) {
			case []interface{}:
				for i, v := range o.([]interface{}) {
					sub, err := resolver(v, append(here, fmt.Sprintf("%d", i)), path, pos+1)
					if err != nil {
						return nil, err
					}
					paths = append(paths, sub...)
				}

			case map[string]interface{}:
				for k, v := range o.(map[string]interface{}) {
					sub, err := resolver(v, append(here, k), path, pos+1)
					if err != nil {
						return nil, err
					}
					paths = append(paths, sub...)
				}

			case map[interface{}]interface{}:
				for k, v := range o.(map[interface{}]interface{}) {
					sub, err := resolver(v, append(here, k.(string)), path, pos+1)
					if err != nil {
						return nil, err
					}
					paths = append(paths, sub...)
				}

			default:
				return nil, TypeMismatchError{
					Path:   path,
					Wanted: "a map or a list",
					Got:    "a scalar",
				}
			}

		} else {
			switch o.(type) {
			case []interface{}:
				i, err := strconv.ParseUint(k, 10, 0)
				if err == nil {
					// if k is an integer (in string form), go by index
					if int(i) >= len(o.([]interface{})) {
						return nil, NotFoundError{
							Path: path[0 : pos+1],
						}
					}
					return resolver(o.([]interface{})[i], append(here, k), path, pos+1)
				}

				// if k is a string, look for immediate map descendants who have
				//     'name', 'key' or 'id' fields matching k
				var found bool
				o, _, found = listFind(o.([]interface{}), []string{"name", "key", "id"}, k)
				if !found {
					return nil, NotFoundError{
						Path: path[0 : pos+1],
					}
				}
				return resolver(o, append(here, k), path, pos+1)

			case map[string]interface{}:
				v, ok := o.(map[string]interface{})[k]
				if !ok {
					return nil, NotFoundError{
						Path: path[0 : pos+1],
					}
				}
				return resolver(v, append(here, k), path, pos+1)

			case map[interface{}]interface{}:
				v, ok := o.(map[interface{}]interface{})[k]
				if !ok {
					return nil, NotFoundError{
						Path: path[0 : pos+1],
					}
				}
				return resolver(v, append(here, k), path, pos+1)

			default:
				return nil, TypeMismatchError{
					Path:   path[0:pos],
					Wanted: "a map or a list",
					Got:    "a scalar",
				}
			}
		}

		return paths, nil
	}

	var path []string
	for _, s := range c.Nodes {
		path = append(path, s)
	}

	l, err := resolver(tree, []string{}, path, 0)
	if err != nil {
		return nil, err
	}

	cursors := []*Cursor{}
	for _, c := range l {
		cursors = append(cursors, c.(*Cursor))
	}
	return cursors, nil
}

// Resolve ...
func (c *Cursor) Resolve(o interface{}) (interface{}, error) {
	var path []string

	for _, k := range c.Nodes {
		path = append(path, k)

		switch o.(type) {
		case map[string]interface{}:
			v, ok := o.(map[string]interface{})[k]
			if !ok {
				return nil, NotFoundError{
					Path: path,
				}
			}
			o = v

		case map[interface{}]interface{}:
			v, ok := o.(map[interface{}]interface{})[k]
			if !ok {
				return nil, NotFoundError{
					Path: path,
				}
			}
			o = v

		case []interface{}:
			i, err := strconv.ParseUint(k, 10, 0)
			if err == nil {
				// if k is an integer (in string form), go by index
				if int(i) >= len(o.([]interface{})) {
					return nil, NotFoundError{
						Path: path,
					}
				}
				o = o.([]interface{})[i]
				continue
			}

			// if k is a string, look for immediate map descendants who have
			//     'name', 'key' or 'id' fields matching k
			var found bool
			o, _, found = listFind(o.([]interface{}), []string{"name", "key", "id"}, k)
			if !found {
				return nil, NotFoundError{
					Path: path,
				}
			}

		default:
			path = path[0 : len(path)-1]
			return nil, TypeMismatchError{
				Path:   path,
				Wanted: "a map or a list",
				Got:    "a scalar",
				Value:  o,
			}
		}
	}

	return o, nil
}

// ResolveString ...
func (c *Cursor) ResolveString(tree interface{}) (string, error) {
	o, err := c.Resolve(tree)
	if err != nil {
		return "", err
	}

	switch o.(type) {
	case string:
		return o.(string), nil
	case int:
		return fmt.Sprintf("%d", o.(int)), nil
	}
	return "", TypeMismatchError{
		Path:   c.Nodes,
		Wanted: "a string",
	}
}
