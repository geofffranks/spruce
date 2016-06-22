package spruce

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/jhunt/ansi"

	. "github.com/geofffranks/spruce/log"
)

// Merger ...
type Merger struct {
	AppendByDefault bool

	Errors MultiError
	depth  int
}

// Merge ...
func Merge(l ...map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	m := &Merger{}
	root := map[interface{}]interface{}{}
	for _, next := range l {
		m.Merge(root, next)
	}
	return root, m.Error()
}

// Error ...
func (m *Merger) Error() error {
	if m.Errors.Count() > 0 {
		return m.Errors
	}
	return nil
}

func deepCopy(orig interface{}) interface{} {
	switch orig.(type) {
	case map[interface{}]interface{}:
		x := map[interface{}]interface{}{}
		for k, v := range orig.(map[interface{}]interface{}) {
			x[k] = deepCopy(v)
		}
		return x

	case []interface{}:
		x := make([]interface{}, len(orig.([]interface{})))
		for i, v := range orig.([]interface{}) {
			x[i] = deepCopy(v)
		}
		return x

	default:
		return orig
	}
}

// Merge ...
func (m *Merger) Merge(a map[interface{}]interface{}, b map[interface{}]interface{}) error {
	m.mergeMap(a, b, "$")
	return m.Error()
}

func (m *Merger) mergeMap(orig map[interface{}]interface{}, n map[interface{}]interface{}, node string) {
	mergeRx := regexp.MustCompile(`^\s*\Q((\E\s*merge\s*.*\Q))\E`)
	pruneRx := regexp.MustCompile(`^\s*\Q((\E\s*prune\s*.*\Q))\E`)
	for k, val := range n {
		path := fmt.Sprintf("%s.%v", node, k)
		if s, ok := val.(string); ok {
			if mergeRx.MatchString(s) {
				m.Errors.Append(ansi.Errorf("@m{%s}: @R{inappropriate use of} @c{(( merge ))} @R{operator outside of a list} (this is @G{spruce}, after all)", path))
			} else if pruneRx.MatchString(s) {
				DEBUG("%s: found (( prune )) operator, adding key to prune list", path)
				keysToPrune = append(keysToPrune, path)
				continue
			}
		}
		if _, exists := orig[k]; exists {
			DEBUG("%s: found upstream, merging it", path)
			orig[k] = m.mergeObj(orig[k], val, path)
		} else {
			DEBUG("%s: not found upstream, adding it", path)
			orig[k] = m.mergeObj(nil, deepCopy(val), path)
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

		case nil:
			orig := map[interface{}]interface{}{}
			m.mergeMap(orig, n.(map[interface{}]interface{}), node)
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

		case nil:
			orig := []interface{}{}
			return m.mergeArray(orig, n.([]interface{}), node)

		default:
			if orig == nil {
				DEBUG("%s: performing array merge (original was nil)", node)
				return m.mergeArray([]interface{}{}, n.([]interface{}), node)
			}

			DEBUG("%s: replacing with new data (original was not an array)", node)
			return t
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
			m.Errors.Append(err)
			return nil
		}
		if err := canKeyMergeArray("original", orig, node, key); err != nil {
			m.Errors.Append(err)
			return nil
		}

		return m.mergeArrayByKey(orig, n[1:], node, key)

	} else if should, idx := shouldInsertIntoArrayBasedOnIndex(n); should {
		if idx < 0 || idx > len(orig) {
			m.Errors.Append(ansi.Errorf("@m{%s}: @R{specified insertion index} @c{%d} @R{is out of bounds}", node, idx))
			return nil
		}

		DEBUG("%s: inserting %d new elements to existing array, at index %d", node, len(n)-1, idx)
		return append(orig[0:idx], append(n[1:], orig[idx:]...)...)

	} else if should, relative, key, name := shouldInsertIntoArrayBasedOnName(n); should {
		DEBUG("%s: inserting %d new elements to existing array %s entry '%s: %s'", node, len(n)-1, relative, key, name)

		if err := canKeyMergeArray("new", n[1:], node, key); err != nil {
			m.Errors.Append(err)
			return nil
		}

		if err := canKeyMergeArray("original", orig, node, key); err != nil {
			m.Errors.Append(err)
			return nil
		}

		idx := getIndexOfEntry(orig, key, name)
		if idx < 0 {
			m.Errors.Append(ansi.Errorf("@m{%s}: @R{unable to find specified insertion point with} @c{'%s: %s'}", node, key, name))
			return nil
		}

		// Since we have a way to identify indiviual entries based on their key/id, we can sanity check for possible duplicates
		for i, entry := range n[1:] {
			obj := entry.(map[interface{}]interface{})
			entryName := obj[key].(string)
			if getIndexOfEntry(orig, key, entryName) > 0 {
				m.Errors.Append(ansi.Errorf("@m{%s}: @R{unable to insert, because new list entry} @c{%d} @R{is detected in both lists} @c{'%s: %s'}", node, i, key, entryName))
				return nil
			}
		}

		if relative == "after" {
			idx++
		}

		return append(orig[0:idx], append(n[1:], orig[idx:]...)...)
	}

	DEBUG("%s: performing index-based array merge", node)
	if err := canKeyMergeArray("original", orig, node, "name"); err == nil {
		if err := canKeyMergeArray("new", n, node, "name"); err == nil {
			return m.mergeArrayByKey(orig, n, node, "name")
		}
	}

	if m.AppendByDefault {
		return append(orig, n...)
	}
	return m.mergeArrayInline(orig, n, node)
}

func (m *Merger) mergeArrayInline(orig []interface{}, n []interface{}, node string) []interface{} {
	length := len(orig)
	if len(n) > len(orig) {
		length = len(n)
	}
	merged := make([]interface{}, length, length)

	var last int
	for i := range orig {
		path := fmt.Sprintf("%s.%d", node, i)
		if i >= len(n) {
			merged[i] = m.mergeObj(nil, orig[i], path)
		} else {
			merged[i] = m.mergeObj(orig[i], n[i], path)
		}
		last = i
	}

	if len(orig) > 0 {
		last++ // move to next index after finishing the orig slice - but only if we looped
	}

	// grab the remainder of n (if any) and append the to the result
	for i := last; i < len(n); i++ {
		path := fmt.Sprintf("%s.%d", node, i)
		DEBUG("%s: appending new data to existing array", path)
		merged[i] = m.mergeObj(nil, n[i], path)
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
		path := fmt.Sprintf("%s.%d", node, i)
		if _, ok := newMap[obj[key]]; ok {
			merged[i] = m.mergeObj(obj, newMap[obj[key]], path)
			delete(newMap, obj[key])
		} else {
			merged[i] = m.mergeObj(nil, obj, path)
		}
	}

	i := 0
	for _, obj := range n {
		obj := obj.(map[interface{}]interface{})
		if _, ok := newMap[obj[key]]; ok {
			path := fmt.Sprintf("%s.%d", node, i)
			DEBUG("%s: appending new data to merged array", path)
			merged = append(merged, m.mergeObj(nil, obj, path))
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

func shouldInsertIntoArrayBasedOnIndex(obj []interface{}) (bool, int) {
	if len(obj) >= 1 && reflect.TypeOf(obj[0]).Kind() == reflect.String {
		re := regexp.MustCompile("^\\Q((\\E\\s*insert\\s+(after|before)\\s+(.+)\\s*\\Q))\\E$")
		if re.MatchString(obj[0].(string)) {
			captures := re.FindStringSubmatch(obj[0].(string))
			relative := strings.TrimSpace(captures[1])
			position := strings.TrimSpace(captures[2])
			if idx, err := strconv.Atoi(position); err == nil {
				// Back out if the index is lower than zero
				if idx < 0 {
					return false, -1
				}

				if relative == "after" {
					return true, idx + 1

				} else {
					return true, idx
				}
			}
		}
	}

	return false, -1
}

func shouldInsertIntoArrayBasedOnName(obj []interface{}) (bool, string, string, string) {
	if len(obj) >= 1 && reflect.TypeOf(obj[0]).Kind() == reflect.String {
		re := regexp.MustCompile("^\\Q((\\E\\s*insert\\s+(after|before)\\s+(on\\s+([^ ]+))?\\s*\"(.+)\"\\s*\\Q))\\E$")
		if re.MatchString(obj[0].(string)) {
			captures := re.FindStringSubmatch(obj[0].(string))

			DEBUG("%d", len(captures))
			for _, entry := range captures {
				DEBUG("%s", entry)
			}

			/* #0 is the whole string,
			 * #1 is after or before
			 * #2 contains the optional 'on <key>' string
			 * #3 contains the optional '<key>' string (because of nested capture groups)
			 * #4 is finally the target "<name>" string
			 */
			if len(captures) == 5 {
				capturedRelative := strings.TrimSpace(captures[1])
				capturedKey := strings.TrimSpace(captures[3])
				capturedName := strings.TrimSpace(captures[4])

				key := "name"
				if capturedKey != "" {
					key = capturedKey
				}

				return true, capturedRelative, key, capturedName
			}
		}
	}

	return false, "", "", ""
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
			return ansi.Errorf("@m{%s.%d}: @R{%s object is a} @c{%s}@R{, not a} @c{map} @R{- cannot merge using keys}", node, i, disp, reflect.TypeOf(o).Kind().String())
		}

		obj := o.(map[interface{}]interface{})
		if _, ok := obj[key]; !ok {
			return ansi.Errorf("@m{%s.%d}: @R{%s object does not contain the key} @c{'%s'}@R{ - cannot merge}", node, i, disp, key)
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

func getIndexOfEntry(list []interface{}, key string, name string) int {
	for i, entry := range list {
		if reflect.TypeOf(entry).Kind() == reflect.Map {
			obj := entry.(map[interface{}]interface{})
			if obj[key] == name {
				return i
			}
		}
	}

	return -1
}
