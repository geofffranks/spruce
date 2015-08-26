package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// NoResolveRequestedError is a custom error indicating no post-process resolution is needed
type NoResolveRequestedError string

func (e *NoResolveRequestedError) Error() string {
	return fmt.Sprintf("%s: does not need to be resolved", string(*e))
}

func postProcessMap(m map[interface{}]interface{}, root map[interface{}]interface{}, node string) error {
	if node == "" {
		node = "$"
	}
	for k, v := range m {
		path := fmt.Sprintf("%s.%v", node, k)
		newVal, err := postProcessObj(v, root, path)
		if err != nil {
			if _, ok := err.(*NoResolveRequestedError); ok {
				continue
			} else {
				return err
			}
		}
		m[k] = newVal
	}
	DEBUG("%s: done post-processing", node)
	return nil
}

func postProcessArray(a []interface{}, root map[interface{}]interface{}, node string) error {
	for i, e := range a {
		path := fmt.Sprintf("%s.[%d]", node, i)
		newVal, err := postProcessObj(e, root, path)
		if err != nil {
			if _, ok := err.(*NoResolveRequestedError); ok {
				continue
			} else {
				return err
			}
		}
		a[i] = newVal
	}
	DEBUG("%s: done post-processing", node)
	return nil
}

func postProcessObj(o interface{}, root map[interface{}]interface{}, node string) (interface{}, error) {
	switch o.(type) {
	case string:
		findPath, should := shouldResolveString(o.(string))
		if should {
			DEBUG("%s: resolving from %s", node, findPath)
			newVal, err := resolveNode(findPath, root)
			if err != nil {
				return nil, fmt.Errorf("%s: Unable to resolve `%s`: `%s", node, findPath, err.Error())
			}
			DEBUG("%s: setting to %#v", node, newVal)
			var retVal interface{}
			if newVal != nil && reflect.TypeOf(newVal).Kind() == reflect.Map {
				retVal = make(map[interface{}]interface{})
				deepCopy(retVal, newVal)
			} else {
				retVal = newVal
			}
			return retVal, nil
		}
		err := NoResolveRequestedError(node)
		return nil, &err
	case map[interface{}]interface{}:
		DEBUG("%s: scanning for values needing to be resolved", node)
		if err := postProcessMap(o.(map[interface{}]interface{}), root, node); err != nil {
			return nil, err
		}
	case []interface{}:
		DEBUG("%s: scanning for values needing to be resolved", node)
		if err := postProcessArray(o.([]interface{}), root, node); err != nil {
			return nil, err
		}
	default:
		DEBUG("%s: does not need to be resolved", node)
		err := NoResolveRequestedError(node)
		return nil, &err
	}
	return o, nil
}

func shouldResolveString(s string) (string, bool) {
	re := regexp.MustCompile("^\\Q((\\E\\s*grab\\s+(\\S+?)\\s*\\Q))\\E$")
	if re.MatchString(s) {
		keys := re.FindStringSubmatch(s)
		if keys[1] != "" {
			return keys[1], true
		}
	}
	return "", false
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

func deepCopy(dst, src interface{}) {
	dv, sv := reflect.ValueOf(dst), reflect.ValueOf(src)

	for _, k := range sv.MapKeys() {
		dv.SetMapIndex(k, sv.MapIndex(k))
	}
}
