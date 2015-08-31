package main

import (
	"fmt"
	"reflect"
)

// PostProcessor interface to allow for flexible post-processing of the tree
type PostProcessor interface {
	PostProcess(interface{}, string) (interface{}, string, error)
}

// Current recursion depth
var CurrentDepth = 0

// Maximum recursion depth
var MaxDepth = 32

func deepCopy(dst, src interface{}) {
	dv, sv := reflect.ValueOf(dst), reflect.ValueOf(src)

	for _, k := range sv.MapKeys() {
		dv.SetMapIndex(k, sv.MapIndex(k))
	}
}

func walkTree(root interface{}, p PostProcessor, node string) error {
	if node == "" {
		node = "$"
		CurrentDepth = 0
	}

	if CurrentDepth >= MaxDepth {
		return fmt.Errorf("%s: hit max recursion depth. You seem to have a self-referencing dataset", node)
	}
	CurrentDepth++

	switch root.(type) {
	case map[interface{}]interface{}:
		for k, v := range root.(map[interface{}]interface{}) {
			path := fmt.Sprintf("%s.%v", node, k)
			val, action, err := p.PostProcess(v, path)
			if err != nil {
				return err
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

			err = walkTree(root.(map[interface{}]interface{})[k], p, path)
			if err != nil {
				return err
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
				return err
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
			err = walkTree(root.([]interface{})[i], p, path)
			if err != nil {
				return err
			}
		}
	}
	CurrentDepth--
	return nil
}
