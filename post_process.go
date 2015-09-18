package main

import (
	"fmt"
	"reflect"
)

// PostProcessor interface to allow for flexible post-processing of the tree
type PostProcessor interface {
	PostProcess(interface{}, string) (interface{}, string, error)
}

func deepCopy(dst, src interface{}) {
	dv, sv := reflect.ValueOf(dst), reflect.ValueOf(src)

	for _, k := range sv.MapKeys() {
		dv.SetMapIndex(k, sv.MapIndex(k))
	}
}

func (m *Merger) Visit(root interface{}, p PostProcessor) {
	m.visit(root, p, "$", 32)
}

func (m *Merger) visit(root interface{}, p PostProcessor, node string, depth int) bool {
	DEBUG("visit(root, p, %q, %d)", node, depth)
	if depth == 0 {
		m.Errors.Push(fmt.Errorf("%s: hit max recursion depth. You seem to have a self-referencing dataset", node))
		DEBUG("cycle detected!")
		return false
	}

	switch root.(type) {
	case map[interface{}]interface{}:
		for k, v := range root.(map[interface{}]interface{}) {
			path := fmt.Sprintf("%s.%v", node, k)

			val, action, err := p.PostProcess(v, path)
			if err != nil {
				m.Errors.Push(err)
				continue
			}

			if action == "replace" {
				var replacement interface{}
				if val != nil && reflect.TypeOf(val).Kind() == reflect.Map {
					replacement = make(map[interface{}]interface{})
					deepCopy(replacement, val)
				} else {
					replacement = val
				}
				root.(map[interface{}]interface{})[k] = replacement
			}

			if ok := m.visit(root.(map[interface{}]interface{})[k], p, path, depth-1); !ok {
				return false
			}
		}
	case []interface{}:
		for i, e := range root.([]interface{}) {
			path := fmt.Sprintf("%s.[%d]", node, i)
			if eMap, ok := e.(map[interface{}]interface{}); ok {
				if name, ok := eMap["name"]; ok {
					path = fmt.Sprintf("%s.%s", node, name)
				}
			}

			val, action, err := p.PostProcess(e, path)
			if err != nil {
				m.Errors.Push(err)
				return true
			}

			if action == "replace" {
				var replacement interface{}
				if val != nil && reflect.TypeOf(val).Kind() == reflect.Map {
					replacement = make(map[interface{}]interface{})
					deepCopy(replacement, val)
				} else {
					replacement = val
				}
				root.([]interface{})[i] = replacement
			}
			if ok := m.visit(root.([]interface{})[i], p, path, depth-1); !ok {
				return false
			}
		}
	}

	return true
}
