package main

import "fmt"
import "reflect"

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
	if shouldMergeArray(n[0]) {
		merged = append(orig, n[1:]...)
	} else if shouldMergeArray(n[len(n)-1]) {
		merged = append(n[:len(n)-1], orig...)
	} else {
		merged = n
	}
	return merged
}

func shouldMergeArray(obj interface{}) bool {
	return reflect.TypeOf(obj).Kind() == reflect.String && obj.(string) == "(( merge ))"
}
