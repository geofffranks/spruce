package main

import (
	"fmt"
	"reflect"
)

func mergeMap(orig map[interface{}]interface{}, n map[interface{}]interface{}, node string) {
	for k, val := range n {
		key := fmt.Sprintf("%v", k)
		path := node + "." + key
		_, exists := orig[key]
		if exists {
			orig[key] = mergeObj(orig[key], val, path)
		} else {
			orig[key] = val
		}
	}
}

func mergeObj(orig interface{}, n interface{}, node string) interface{} {
	switch t := n.(type) {
	case map[interface{}]interface{}:
		switch orig.(type) {
		case map[interface{}]interface{}:
			mergeMap(orig.(map[interface{}]interface{}), n.(map[interface{}]interface{}), node)
		default:
			orig = t
		}
	case []interface{}:
		switch orig.(type) {
		case []interface{}:
			orig = mergeArray(orig.([]interface{}), n.([]interface{}), node)
		default:
			orig = t
		}

	default:
		orig = t
	}
	return orig
}

func mergeArray(orig []interface{}, n []interface{}, node string) []interface{} {
	var merged []interface{}
	if shouldAppendToArray(n) {
		merged = append(orig, n[1:]...)
	} else if shouldPrependToArray(n) {
		merged = append(n[1:], orig...)
	} else if shouldInlineMergeArray(n) {
		length := len(orig)
		// len(n)-1 accounts for the "(( inline ))" initial element that should be dropped
		if len(n)-1 > len(orig) {
			length = len(n) - 1
		}
		merged = make([]interface{}, length, length)

		var last int
		for i := range orig {
			// i+1 accounts for the "(( inline ))" initial element that should be dropped
			if i+1 >= len(n) {
				merged[i] = orig[i]
			} else {
				merged[i] = mergeObj(orig[i], n[i+1], node)
			}
			last = i
		}

		last++ // move to next index after finishing the orig slice

		// grab the remainder of n (if any), accounting for the "(( inline ))" element
		// and append the to the result
		for i := last; i < len(n)-1; i++ {
			merged[i] = n[i+1]
		}
	} else {
		merged = n
	}
	return merged
}

func shouldInlineMergeArray(obj []interface{}) bool {
	return len(obj) >= 1 && reflect.TypeOf(obj[0]).Kind() == reflect.String && obj[0].(string) == "(( inline ))"
}
func shouldAppendToArray(obj []interface{}) bool {
	return len(obj) >= 1 && reflect.TypeOf(obj[0]).Kind() == reflect.String && obj[0].(string) == "(( append ))"
}
func shouldPrependToArray(obj []interface{}) bool {
	return len(obj) >= 1 && reflect.TypeOf(obj[0]).Kind() == reflect.String && obj[0].(string) == "(( prepend ))"
}
