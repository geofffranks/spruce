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

// Array operation helper structure
type ModificationDefinition struct {
	index int

	key  string
	name string

	delete bool

	relative string
	list     []interface{}
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

	if shouldInlineMergeArray(n) {
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

	} else if should, modificationDefinitions := shouldModifyArray(n); should {
		DEBUG("%s: performing %d modification operations against list", node, len(modificationDefinitions))

		// Create a copy of orig for the (multiple) modifications that are about to happen
		result := make([]interface{}, len(orig))
		copy(result, orig)

		var idx int

		// Process the insert definitions that were found in the new list
		for i := range modificationDefinitions {
			if modificationDefinitions[i].key == "" && modificationDefinitions[i].name == "" { // Index comes directly from operation definition
				idx = modificationDefinitions[i].index

				// Replace the -1 marker with the actual 'end' index of the array
				if idx == -1 {
					idx = len(result)
				}

			} else { // Index look-up based on key and name
				key := modificationDefinitions[i].key
				name := modificationDefinitions[i].name
				delete := modificationDefinitions[i].delete

				// Sanity check original list, list must contain key/id based entries
				if err := canKeyMergeArray("original", result, node, key); err != nil {
					m.Errors.Append(err)
					return nil
				}

				// Sanity check new list, depending on the operation type (delete or insert)
				if delete == false {

				    // Sanity check new list, list must contain key/id based entries
					if err := canKeyMergeArray("new", modificationDefinitions[i].list, node, key); err != nil {
						m.Errors.Append(err)
						return nil
					}

					// Since we have a way to identify indiviual entries based on their key/id, we can sanity check for possible duplicates
					for _, entry := range modificationDefinitions[i].list {
						obj := entry.(map[interface{}]interface{})
						entryName := obj[key].(string)
						if getIndexOfEntry(result, key, entryName) > 0 {
							m.Errors.Append(ansi.Errorf("@m{%s}: @R{unable to insert, because new list entry} @c{'%s: %s'} @R{is detected multiple times}", node, key, entryName))
							return nil
						}
					}
				} else {
				    // Sanity check for delete operation, ensure no orphan entries follow the operator definition					
					if len(modificationDefinitions[i].list) > 0 {
						m.Errors.Append(ansi.Errorf("@m{%s}: @R{unable to delete, orphan entries found after} @c{'%s: %s'}", node, key, name))
						return nil
					}
				}

				// Look up the index of the specified insertion point (based on its key/name)
				idx = getIndexOfEntry(result, key, name)
				if idx < 0 {
					m.Errors.Append(ansi.Errorf("@m{%s}: @R{unable to find specified modification point with} @c{'%s: %s'}", node, key, name))
					return nil
				}
			}

			// If after is specified, add one to the index to actually put the entry where it is expected
			if modificationDefinitions[i].relative == "after" {
				idx++
			}

			// Back out if idx is smaller than 0, or greater than the length (for inserts), or greater/equal than the length (for deletes)
			if (idx < 0) || ( modificationDefinitions[i].delete == false && idx > len(result) )  || ( modificationDefinitions[i].delete == true && idx >= len(result) ) {
				m.Errors.Append(ansi.Errorf("@m{%s}: @R{unable to modify the list, because specified index} @c{%d} @R{is out of bounds}", node, idx))
				return nil
			}

			if modificationDefinitions[i].delete == false {
				DEBUG("%s: inserting %d new elements to existing array at index %d", node, len(modificationDefinitions[i].list), idx)
				result = insertIntoList(result, idx, modificationDefinitions[i].list)
			} else {
				DEBUG("%s: deleting element at array index %d", node, idx)
				result = deleteIndexFromList(result, idx)
			}
		}

		return result
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

func shouldModifyArray(obj []interface{}) (bool, []ModificationDefinition) {
	if len(obj) >= 1 && reflect.TypeOf(obj[0]).Kind() == reflect.String {
		appendRegEx := regexp.MustCompile("^\\Q((\\E\\s*append\\s*\\Q))\\E$")
		prependRegEx := regexp.MustCompile("^\\Q((\\E\\s*prepend\\s*\\Q))\\E$")
		insertByIdxRegEx := regexp.MustCompile("^\\Q((\\E\\s*insert\\s+(after|before)\\s+(\\d+)\\s*\\Q))\\E$")
		insertByNameRegEx := regexp.MustCompile("^\\Q((\\E\\s*insert\\s+(after|before)\\s+([^ ]+)?\\s*\"(.+)\"\\s*\\Q))\\E$")
		deleteByIdxRegEx := regexp.MustCompile("^\\Q((\\E\\s*delete\\s+(\\d+)\\s*\\Q))\\E$")
		deleteByNameRegEx := regexp.MustCompile("^\\Q((\\E\\s*delete\\s+([^ ]+)?\\s*\"(.+)\"\\s*\\Q))\\E$")

		var result []ModificationDefinition
		for i, entry := range obj {
			if reflect.TypeOf(entry).Kind() == reflect.String {
				if appendRegEx.MatchString(entry.(string)) { // check for (( append ))
					result = append(result, ModificationDefinition{index: -1})
					continue

				} else if prependRegEx.MatchString(entry.(string)) { // check for (( prepend ))
					result = append(result, ModificationDefinition{index: 0})
					continue

				} else if insertByIdxRegEx.MatchString(entry.(string)) { // check for (( insert ... <idx> ))
					/* #0 is the whole string,
					 * #1 is after or before
					 * #2 is the insertion index
					 */
					if captures := insertByIdxRegEx.FindStringSubmatch(entry.(string)); len(captures) == 3 {
						relative := strings.TrimSpace(captures[1])
						position := strings.TrimSpace(captures[2])
						if idx, err := strconv.Atoi(position); err == nil {
							result = append(result, ModificationDefinition{index: idx, relative: relative})
							continue
						}
					}

				} else if insertByNameRegEx.MatchString(entry.(string)) { // check for (( insert ... "<name>" ))
					/* #0 is the whole string,
					 * #1 is after or before
					 * #2 contains the optional '<key>' string
					 * #3 is finally the target "<name>" string
					 */
					if captures := insertByNameRegEx.FindStringSubmatch(entry.(string)); len(captures) == 4 {
						relative := strings.TrimSpace(captures[1])
						key := strings.TrimSpace(captures[2])
						name := strings.TrimSpace(captures[3])

						if key == "" {
							key = "name"
						}

						result = append(result, ModificationDefinition{relative: relative, key: key, name: name})
						continue
					}
				} else if deleteByIdxRegEx.MatchString(entry.(string)) { // check for (( delete <idx> ))
					/* #0 is the whole string,
					 * #1 is idx
					 */
					if captures := deleteByIdxRegEx.FindStringSubmatch(entry.(string)); len(captures) == 2 {
						position := strings.TrimSpace(captures[1])
						if idx, err := strconv.Atoi(position); err == nil {
							result = append(result, ModificationDefinition{index: idx, delete: true})
							continue
						}
					}
				} else if deleteByNameRegEx.MatchString(entry.(string)) { // check for (( delete "<name>" ))
					/* #0 is the whole string,
					 * #1 contains the optional '<key>' string
					 * #2 is finally the target "<name>" string
					 */
					if captures := deleteByNameRegEx.FindStringSubmatch(entry.(string)); len(captures) == 3 {
						key := strings.TrimSpace(captures[1])
						name := strings.TrimSpace(captures[2])

						if key == "" {
							key = "name"
						}

						result = append(result, ModificationDefinition{key: key, name: name, delete: true})
						continue
					}
				}
			}

			lastResultIdx := len(result) - 1
			if lastResultIdx >= 0 {
				// Add the current entry to the 'current' modification definition record (gathering the list)
				result[lastResultIdx].list = append(result[lastResultIdx].list, entry)
			} else {
				// Having no last result index at hand means we are dealing with an orphaned list entry
				DEBUG("List entry %d cannot be connected to a modification operation (orphaned entry)", i)
				return false, nil
			}
		}
		if len(result) > 0 {
			return true, result
		}
	}
	return false, nil
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

func insertIntoList(orig []interface{}, idx int, list []interface{}) []interface{} {
	prefix := make([]interface{}, idx)
	copy(prefix, orig[0:idx])

	sublist := make([]interface{}, len(list))
	copy(sublist, list)

	suffix := make([]interface{}, len(orig)-idx)
	copy(suffix, orig[idx:])

	return append(prefix, append(sublist, suffix...)...)
}

func deleteIndexFromList(orig []interface{}, idx int) []interface{} {
	tmp := make([]interface{}, len(orig))
	copy(tmp, orig)

	return append(tmp[:idx], tmp[idx+1:]...)
}
