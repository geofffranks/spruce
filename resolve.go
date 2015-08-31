package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func resolveNode(target string, lookup map[interface{}]interface{}) (interface{}, error) {
	keys := strings.Split(target, ".")

	return resolveNodeObj(keys, lookup)
}

func resolveNodeObj(keys []string, lookup interface{}) (interface{}, error) {
	toFind, keys := keys[0], keys[1:]
	DEBUG("   RESOLVE: searching for %q", toFind)
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
	default:
		return nil, fmt.Errorf("Tried to reference unspported value type '%s'. This is a post-processing bug", reflect.TypeOf(lookup).String())
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
