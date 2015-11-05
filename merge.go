package main

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

type ErrorList struct {
	errs []error
}

func (el ErrorList) Error() string {
	str := make([]string, len(el.errs))
	for i, s := range el.errs {
		str[i] = fmt.Sprintf(" - %s\n", s)
	}
	sort.Strings(str)
	return fmt.Sprintf("%d error(s) detected:\n%s\n", len(el.errs), strings.Join(str, ""))
}

func (el *ErrorList) Count() int {
	return len(el.errs)
}
func (el *ErrorList) Push(e error) {
	if e != nil {
		el.errs = append(el.errs, e)
	}
}

type Merger struct {
	Errors ErrorList
	depth  int
}

func Merge(l ...map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	m := &Merger{}
	root := map[interface{}]interface{}{}
	for _, next := range l {
		m.Merge(root, next)
	}
	return root, m.Error()
}

func (m *Merger) Error() error {
	if m.Errors.Count() > 0 {
		return m.Errors
	}
	return nil
}

func (m *Merger) Merge(a map[interface{}]interface{}, b map[interface{}]interface{}) error {
	m.mergeMap(a, b, "$")
	return m.Error()
}

func (m *Merger) mergeMap(orig map[interface{}]interface{}, n map[interface{}]interface{}, node string) {
	for k, val := range n {
		path := fmt.Sprintf("%s.%v", node, k)
		if _, exists := orig[k]; exists {
			orig[k] = m.mergeObj(orig[k], val, path)
		} else {
			DEBUG("%s: not found upstream, adding it", path)
			orig[k] = val
		}
	}
}

func (m *Merger) mergeObj(orig interface{}, n interface{}, node string) interface{} {
	switch t := n.(type) {
	case map[interface{}]interface{}:
		switch orig.(type) {
		case map[interface{}]interface{}:
			DEBUG("%s: performing map merge", node)
			m.mergeMap(orig.(map[interface{}]interface{}), n.(map[interface{}]interface{}), node)
			return orig

		default:
			DEBUG("%s: replacing with new data (original was not a map)", node)
			return t
		}

	case []interface{}:
		switch orig.(type) {
		case []interface{}:
			DEBUG("%s: performing array merge", node)
			return m.mergeArray(orig.([]interface{}), n.([]interface{}), node)

		default:
			if orig == nil {
				DEBUG("%s: performing array merge (original was nil)", node)
				return m.mergeArray([]interface{}{}, n.([]interface{}), node)

			} else {
				DEBUG("%s: replacing with new data (original was not an array)", node)
				return t
			}
		}

	default:
		DEBUG("%s: replacing with new data (new data is neither map nor array)", node)
		return t
	}
}

func (m *Merger) mergeArray(orig []interface{}, n []interface{}, node string) []interface{} {

	if shouldAppendToArray(n) {
		DEBUG("%s: appending %d new elements to existing array, starting at index %d", node, len(n)-1, len(orig))
		return append(orig, n[1:]...)

	} else if shouldPrependToArray(n) {
		DEBUG("%s: prepending %d new elements to existing array", node, len(n)-1)
		return append(n[1:], orig...)

	} else if shouldInlineMergeArray(n) {
		DEBUG("%s: performing explicit inline array merge", node)
		return m.mergeArrayInline(orig, n[1:], node)

	} else if shouldReplaceArray(n) {
		DEBUG("%s: replacing with new data", node)
		return n[1:]

	} else if should, key := shouldKeyMergeArray(n); should {
		DEBUG("%s: performing key-based array merge, using key '%s'", node, key)

		if err := canKeyMergeArray("new", n[1:], node, key); err != nil {
			m.Errors.Push(err)
			return nil
		}
		if err := canKeyMergeArray("original", orig, node, key); err != nil {
			m.Errors.Push(err)
			return nil
		}

		return m.mergeArrayByKey(orig, n[1:], node, key)

	} else {
		DEBUG("%s: performing index-based array merge", node)

		if err := canKeyMergeArray("original", orig, node, "name"); err == nil {
			if err := canKeyMergeArray("new", n, node, "name"); err == nil {
				return m.mergeArrayByKey(orig, n, node, "name")
			}
		}

		return m.mergeArrayInline(orig, n, node)
	}
}

func (m *Merger) mergeArrayInline(orig []interface{}, n []interface{}, node string) []interface{} {
	length := len(orig)
	if len(n) > len(orig) {
		length = len(n)
	}
	merged := make([]interface{}, length, length)

	var last int
	for i := range orig {
		if i >= len(n) {
			merged[i] = orig[i]
		} else {
			merged[i] = m.mergeObj(orig[i], n[i], fmt.Sprintf("%s.%d", node, i))
		}
		last = i
	}

	if len(orig) > 0 {
		last++ // move to next index after finishing the orig slice - but only if we looped
	}

	// grab the remainder of n (if any) and append the to the result
	for i := last; i < len(n); i++ {
		DEBUG("%s.%d: appending new data to existing array", node, i)
		merged[i] = m.mergeObj(nil, n[i], fmt.Sprintf("%s.%d", node, i))
	}

	return merged
}

func (m *Merger) mergeArrayByKey(orig []interface{}, n []interface{}, node string, key string) []interface{} {
	merged := make([]interface{}, len(orig), len(orig))
	newMap := make(map[interface{}]interface{})
	for _, o := range n {
		obj := o.(map[interface{}]interface{})
		newMap[obj[key]] = obj
	}
	for i, o := range orig {
		obj := o.(map[interface{}]interface{})
		if _, ok := newMap[obj[key]]; ok {
			merged[i] = m.mergeObj(obj, newMap[obj[key]], fmt.Sprintf("%s.%d", node, i))
			delete(newMap, obj[key])
		} else {
			merged[i] = obj
		}
	}

	i := 0
	for _, obj := range n {
		obj := obj.(map[interface{}]interface{})
		if _, ok := newMap[obj[key]]; ok {
			DEBUG("%s.%d: appending new data to merged array", node, i)
			merged = append(merged, obj)
			i++
		}
	}

	return merged
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

func canKeyMergeArray(disp string, array []interface{}, node string, key string) error {
	// ensure that all elements of `array` are maps,
	// and that they contain the key `key`

	for i, o := range array {
		if reflect.TypeOf(o).Kind() != reflect.Map {
			return fmt.Errorf("%s.%d: %s object is a %s, not a map - cannot merge using keys", node, i, disp, reflect.TypeOf(o).Kind().String())
		}

		obj := o.(map[interface{}]interface{})
		if _, ok := obj[key]; !ok {
			return fmt.Errorf("%s.%d: %s object does not contain the key '%s' - cannot merge", node, i, disp, key)
		}
	}
	return nil
}

func shouldReplaceArray(obj []interface{}) bool {
	if len(obj) >= 1 && reflect.TypeOf(obj[0]).Kind() == reflect.String {
		re := regexp.MustCompile(`^\Q((\E\s*replace\s*\Q))\E$`)

		if re.MatchString(obj[0].(string)) {
			return true
		}
	}
	return false
}
