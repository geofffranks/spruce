package main

import (
	"fmt"
	"reflect"
)

func mergeMap(orig map[interface{}]interface{}, n map[interface{}]interface{}, node string) {
	if node == "" {
		node = "$"
	}

	for k, val := range n {
		path := fmt.Sprintf("%s.%v", node, k)
		_, exists := orig[k]
		if exists {
			orig[k] = mergeObj(orig[k], val, path)
		} else {
			DEBUG("%s: not found upstream, adding it", path)
			orig[k] = val
		}
	}
}

func mergeObj(orig interface{}, n interface{}, node string) interface{} {
	switch t := n.(type) {
	case map[interface{}]interface{}:
		switch orig.(type) {
		case map[interface{}]interface{}:
			DEBUG("%s: performing map merge", node)
			mergeMap(orig.(map[interface{}]interface{}), n.(map[interface{}]interface{}), node)
		default:
			DEBUG("%s: replacing with new data (original was not a map)", node)
			orig = t
		}
	case []interface{}:
		switch orig.(type) {
		case []interface{}:
			DEBUG("%s: performing array merge", node)
			orig = mergeArray(orig.([]interface{}), n.([]interface{}), node)
		default:
			DEBUG("%s: replacing with new data (original was not an array)", node)
			orig = t
		}

	default:
		DEBUG("%s: replacing with new data (new data is neither map nor array)", node)
		orig = t
	}
	return orig
}

func mergeArray(orig []interface{}, n []interface{}, node string) []interface{} {
	var merged []interface{}
	if shouldAppendToArray(n) {
		DEBUG("%s: appending %d new elements to existing array, starting at index %d", node, len(n)-1, len(orig))
		merged = append(orig, n[1:]...)
	} else if shouldPrependToArray(n) {
		DEBUG("%s: prepending %d new elements to existing array", node, len(n)-1)
		merged = append(n[1:], orig...)
	} else if shouldInlineMergeArray(n) {
		DEBUG("%s: performing inline array merge", node)
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
				merged[i] = mergeObj(orig[i], n[i+1], fmt.Sprintf("%s.%d", node, i))
			}
			last = i
		}

		last++ // move to next index after finishing the orig slice

		// grab the remainder of n (if any), accounting for the "(( inline ))" element
		// and append the to the result
		for i := last; i < len(n)-1; i++ {
			DEBUG("%s.%d: appending new data to existing array", node, i)
			merged[i] = n[i+1]
		}
	} else {
		DEBUG("%s: replacing with new data (no specific array merge behavior requested)", node)
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
