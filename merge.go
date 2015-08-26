package main

import (
	"fmt"
	"reflect"
	"regexp"
)

func mergeMap(orig map[interface{}]interface{}, n map[interface{}]interface{}, node string) error {
	if node == "" {
		node = "$"
	}

	for k, val := range n {
		path := fmt.Sprintf("%s.%v", node, k)
		_, exists := orig[k]
		if exists {
			// check the value of the object to see if it is '(( prune ))'
			if reflect.TypeOf(val).Kind() == reflect.String {
				re := regexp.MustCompile(`\Q((\E\s*prune\s*\Q))\E`)
				if re.MatchString(val.(string)) {
					delete(orig, k)
					continue
				}
			}
			o, err := mergeObj(orig[k], val, path)
			if err != nil {
				return err
			}
			orig[k] = o
		} else {
			DEBUG("%s: not found upstream, adding it", path)
			orig[k] = val
		}
	}
	return nil
}

func mergeObj(orig interface{}, n interface{}, node string) (interface{}, error) {
	switch t := n.(type) {
	case map[interface{}]interface{}:
		switch orig.(type) {
		case map[interface{}]interface{}:
			DEBUG("%s: performing map merge", node)
			err := mergeMap(orig.(map[interface{}]interface{}), n.(map[interface{}]interface{}), node)
			if err != nil {
				return nil, err
			}
		default:
			DEBUG("%s: replacing with new data (original was not a map)", node)
			orig = t
		}
	case []interface{}:
		switch orig.(type) {
		case []interface{}:
			DEBUG("%s: performing array merge", node)
			a, err := mergeArray(orig.([]interface{}), n.([]interface{}), node)
			if err != nil {
				return nil, err
			}
			orig = a
		default:
			if orig == nil {
				DEBUG("%s: performing array merge (original was nil)", node)
				a, err := mergeArray([]interface{}{}, n.([]interface{}), node)
				if err != nil {
					return nil, err
				}
				orig = a
			} else {
				DEBUG("%s: replacing with new data (original was not an array)", node)
				orig = t
			}
		}

	default:
		DEBUG("%s: replacing with new data (new data is neither map nor array)", node)
		orig = t
	}
	return orig, nil
}

func mergeArray(orig []interface{}, n []interface{}, node string) ([]interface{}, error) {
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
				o, err := mergeObj(orig[i], n[i+1], fmt.Sprintf("%s.%d", node, i))
				if err != nil {
					return nil, err
				}
				merged[i] = o
			}
			last = i
		}

		if len(orig) > 0 {
			last++ // move to next index after finishing the orig slice - but only if we looped
		}

		// grab the remainder of n (if any), accounting for the "(( inline ))" element
		// and append the to the result
		for i := last; i < len(n)-1; i++ {
			DEBUG("%s.%d: appending new data to existing array", node, i)
			o, err := mergeObj(nil, n[i+1], fmt.Sprintf("%s.%d", node, i))
			if err != nil {
				return nil, err
			}
			merged[i] = o
		}
	} else if should, key := shouldKeyMergeArray(n); should {
		DEBUG("%s: performing key-based array merge, using key '%s'", node, key)
		merged = make([]interface{}, len(orig), len(orig))

		newArray := n[1:]

		newMap := make(map[interface{}]interface{})
		for i, o := range newArray {
			if reflect.TypeOf(o).Kind() != reflect.Map {
				return nil, fmt.Errorf("%s.%d: new object is a %s, not a map - cannot merge using keys", node, i, reflect.TypeOf(o).Kind().String())
			}
			obj := o.(map[interface{}]interface{})
			if _, ok := obj[key]; !ok {
				return nil, fmt.Errorf("%s.%d: new object does not contain the key '%s' - cannot merge", node, i, key)
			}

			newMap[obj[key]] = obj
		}

		for i, obj := range orig {
			if reflect.TypeOf(obj).Kind() != reflect.Map {
				return nil, fmt.Errorf("%s.%d: original object is a %s, not a map - cannot merge using keys", node, i, reflect.TypeOf(obj).Kind().String())
			}
			obj := obj.(map[interface{}]interface{})
			if _, ok := obj[key]; !ok {
				return nil, fmt.Errorf("%s.%d: original object does not contain the key '%s' - cannot merge", node, i, key)
			}

			if _, ok := newMap[obj[key]]; ok {
				o, err := mergeObj(obj, newMap[obj[key]], fmt.Sprintf("%s.%d", node, i))
				if err != nil {
					return nil, err
				}
				merged[i] = o
				delete(newMap, obj[key])
			} else {
				merged[i] = obj
			}
		}

		i := 0
		for _, obj := range newArray {
			obj := obj.(map[interface{}]interface{})
			if _, ok := newMap[obj[key]]; ok {
				DEBUG("%s.%d: appending new data to merged array", node, i)
				merged = append(merged, obj)
				i++
			}
		}
	} else {
		DEBUG("%s: replacing with new data (no specific array merge behavior requested)", node)
		merged = n
	}
	return merged, nil
}

func shouldInlineMergeArray(obj []interface{}) bool {
	if len(obj) >= 1 && reflect.TypeOf(obj[0]).Kind() == reflect.String {
		re := regexp.MustCompile("^\\Q((\\E\\s*inline\\s*\\Q))\\E$")
		if re.MatchString(obj[0].(string)) {
			return true
		}
	}
	return false
}
func shouldAppendToArray(obj []interface{}) bool {
	if len(obj) >= 1 && reflect.TypeOf(obj[0]).Kind() == reflect.String {
		re := regexp.MustCompile("^\\Q((\\E\\s*append\\s*\\Q))\\E$")
		if re.MatchString(obj[0].(string)) {
			return true
		}
	}
	return false
}
func shouldPrependToArray(obj []interface{}) bool {
	if len(obj) >= 1 && reflect.TypeOf(obj[0]).Kind() == reflect.String {
		re := regexp.MustCompile("^\\Q((\\E\\s*prepend\\s*\\Q))\\E$")
		if re.MatchString(obj[0].(string)) {
			return true
		}
	}
	return false
}
func shouldKeyMergeArray(obj []interface{}) (bool, string) {
	key := "name"

	if len(obj) >= 1 && reflect.TypeOf(obj[0]).Kind() == reflect.String {
		re := regexp.MustCompile("^\\Q((\\E\\s*merge(?:\\s+on\\s+(.*?))?\\s*\\Q))\\E$")

		if re.MatchString(obj[0].(string)) {
			keys := re.FindStringSubmatch(obj[0].(string))
			if keys[1] != "" {
				key = keys[1]
			}
			return true, key
		}
	}
	return false, ""
}
