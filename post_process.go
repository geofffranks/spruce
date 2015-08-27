package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type PostProcessor interface {
	PostProcess(interface{}, string) (interface{}, string, error)
	Action() string
}

var CURRENT_DEPTH = 0
var MAX_DEPTH = 32

func deepCopy(dst, src interface{}) {
	dv, sv := reflect.ValueOf(dst), reflect.ValueOf(src)

	for _, k := range sv.MapKeys() {
		dv.SetMapIndex(k, sv.MapIndex(k))
	}
}

func walkTree(root interface{}, p PostProcessor, node string) error {
	if node == "" {
		node = "$"
		CURRENT_DEPTH = 0
	}

	if CURRENT_DEPTH >= MAX_DEPTH {
		return fmt.Errorf("%s: hit max recursion depth. You seem to have a self-referencing dataset.", node)
	}
	CURRENT_DEPTH++

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

			DEBUG("%s: scanning for values to %s", path, p.Action())
			err = walkTree(root.(map[interface{}]interface{})[k], p, path)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		for i, e := range root.([]interface{}) {
			path := fmt.Sprintf("%s.[%d]", node, i)
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
			DEBUG("%s: scanning for values needing to be resolved", path)
			err = walkTree(root.([]interface{})[i], p, path)
			if err != nil {
				return err
			}
		}
	}
	CURRENT_DEPTH--
	return nil
}

type DeReferencer struct {
	root map[interface{}]interface{}
}

func (d DeReferencer) Action() string {
	return "dereference"
}

func (d DeReferencer) PostProcess(o interface{}, node string) (interface{}, string, error) {
	if reflect.TypeOf(o).Kind() == reflect.String {
		re := regexp.MustCompile("^\\Q((\\E\\s*grab\\s+(\\S+?)\\s*\\Q))\\E$")
		if re.MatchString(o.(string)) {
			keys := re.FindStringSubmatch(o.(string))
			if keys[1] != "" {
				target := keys[1]
				DEBUG("%s: resolving from %s", node, target)
				val, err := resolveNode(target, d.root)
				if err != nil {
					return nil, "error", fmt.Errorf("%s: Unable to resolve `%s`: `%s", node, target, err.Error())
				}
				DEBUG("%s: setting to %#v", node, val)
				return val, "replace", nil
			}
		}
	}

	return nil, "ignore", nil
}

func resolveNode(target string, lookup map[interface{}]interface{}) (interface{}, error) {
	keys := strings.Split(target, ".")

	return resolveNodeObj(keys, lookup)
}

func resolveNodeObj(keys []string, lookup interface{}) (interface{}, error) {
	toFind, keys := keys[0], keys[1:]
	DEBUG("   RESOLVE: searching for %s", toFind)
	switch lookup.(type) {
	case map[interface{}]interface{}:
		m := lookup.(map[interface{}]interface{})
		val, ok := m[toFind]
		if ok {
			return recursiveResolve(toFind, keys, val)
		}
	case []interface{}:
		a := lookup.([]interface{})
		re := regexp.MustCompile("^\\Q[\\E(\\d+)\\Q]\\E$")
		if re.MatchString(toFind) {
			index, err := strconv.Atoi(re.FindStringSubmatch(toFind)[1])
			if err != nil {
				return nil, fmt.Errorf("%s` somehow spruce detected a numeric index, but couldn't convert it to an integer", toFind)
			}
			if index >= len(a) {
				return nil, fmt.Errorf("%s` array's highest index is %d", toFind, (len(a) - 1))
			}
			return recursiveResolve(toFind, keys, a[index])
		}
		for _, o := range a {
			if reflect.TypeOf(o).Kind() == reflect.Map {
				m := o.(map[interface{}]interface{})
				name, ok := m["name"]
				if ok && name == toFind {
					return recursiveResolve(toFind, keys, o)
				}
			}
		}
	}
	return nil, fmt.Errorf("%s` could not be found in the YAML datastructure", toFind)
}

func recursiveResolve(current string, keys []string, lookup interface{}) (interface{}, error) {
	if len(keys) > 0 {
		if reflect.TypeOf(lookup).Kind() != reflect.Map && reflect.TypeOf(lookup).Kind() != reflect.Slice {
			return nil, fmt.Errorf("%s` has no sub-objects", current)
		}
		val, err := resolveNodeObj(keys, lookup)
		if err != nil {
			return nil, fmt.Errorf("%s.%s", current, err.Error())
		}
		return val, nil
	}
	return lookup, nil
}
