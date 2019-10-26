package spruce

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/starkandwayne/goutils/ansi"

	. "github.com/geofffranks/spruce/log"
)

type listOp int

const (
	listOpMergeDefault listOp = iota
	listOpMergeOnKey
	listOpMergeInline
	listOpReplace
	listOpInsert
	listOpDelete
)

// Merger ...
type Merger struct {
	AppendByDefault bool

	Errors MultiError
	depth  int
}

// ModificationDefinition encapsulates the details of an array modification:
// (1) the type of modification, e.g. insert, delete, replace
// (2) an optional guide to the specific part of the array to be modified,
//    for example the index at which an insertion should be done
// (3) an optional list of entries to be added or merged into the array
type ModificationDefinition struct {
	listOp listOp

	index    int
	key      string
	name     string
	relative string

	list []interface{}
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

func getDefaultIdentifierKey() string {
	// Use environment variable override, if set
	if os.Getenv("DEFAULT_ARRAY_MERGE_KEY") != "" {
		return os.Getenv("DEFAULT_ARRAY_MERGE_KEY")
	}

	// the built-in default: name
	return "name"
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
	for k, val := range n {
		path := fmt.Sprintf("%s.%v", node, k)
		if s, ok := val.(string); ok && mergeRx.MatchString(s) {
			m.Errors.Append(ansi.Errorf("@m{%s}: @R{inappropriate use of} @c{(( merge ))} @R{operator outside of a list} (this is @G{spruce}, after all)", path))
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
	// regular expression to search for prune and sort operator to make their
	// special behavior possible
	pruneRx := regexp.MustCompile(`^\s*\Q((\E\s*prune\s*\Q))\E`)
	sortRx := regexp.MustCompile(`^\s*\Q((\E\s*sort(?:\s+by\s+(.*?))?\s*\Q))\E$`)

	// prune/sort operator special behavior I:
	// operator is defined in the original object and will now be overwritten by
	// the new value. Therefore, remember that the operator was here at that path
	//
	// prune/sort operator special behavior II:
	// operator is defined in the new object and would therefore overwrite the
	// original content. In this case, keep the original content as it is and mark
	// that the operator occurred at this path
	//
	// common requirement is that both original and new object values are strings
	origString, origOk := orig.(string)
	newString, newOk := n.(string)
	switch {
	case origOk && pruneRx.MatchString(origString):
		DEBUG("%s: a (( prune )) operator is about to be replaced, check if its path needs to be saved", node)
		addToPruneListIfNecessary(strings.Replace(node, "$.", "", -1))

	case newOk && pruneRx.MatchString(newString) && orig != nil:
		DEBUG("%s: a (( prune )) operator is about to replace existing content, check if its path needs to be saved", node)
		addToPruneListIfNecessary(strings.Replace(node, "$.", "", -1))
		return orig

	case origOk && sortRx.MatchString(origString):
		DEBUG("%s: a (( sort )) operator is about to be replaced, check if its path needs to be saved", node)
		addToSortListIfNecessary(origString, strings.Replace(node, "$.", "", -1))

	case newOk && sortRx.MatchString(newString) && orig != nil:
		DEBUG("%s: a (( sort )) operator is about to replace existing content, check if its path needs to be saved", node)
		addToSortListIfNecessary(newString, strings.Replace(node, "$.", "", -1))
		return orig
	}

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
	modificationDefinitions := getArrayModifications(n, isSimpleList(orig))
	DEBUG("%s: performing %d modification operations against list", node, len(modificationDefinitions))

	// Create a copy of orig for the (multiple) modifications that are about to happen
	result := make([]interface{}, len(orig))
	copy(result, orig)

	// Process the modifications definitions that were found in the new list
	for i, modificationDefinition := range modificationDefinitions {
		DEBUG("  #%d %#v", i, modificationDefinition)

		// insert/delete operations will use a list index later in this loop block
		var idx int

		// Special tag for default behavior. Cannot be invoked explicitly by users
		if modificationDefinition.listOp == listOpMergeDefault {
			result = m.mergeArrayDefault(orig, modificationDefinitions[0].list, node)
			continue
		}

		// Perform a merge on key list modification (merge in new list on original)
		if modificationDefinition.listOp == listOpMergeOnKey {
			key := modificationDefinition.key
			if key == "" {
				key = getDefaultIdentifierKey()
			}

			if err := canKeyMergeArray("new", modificationDefinition.list, node, key); err != nil {
				m.Errors.Append(err)
				return nil
			}
			if err := canKeyMergeArray("original", orig, node, key); err != nil {
				m.Errors.Append(err)
				return nil
			}

			result = m.mergeArrayByKey(result, modificationDefinition.list, node, key)
			continue
		}

		// Perform a merge inline list modification
		if modificationDefinition.listOp == listOpMergeInline {
			result = m.mergeArrayInline(result, modificationDefinition.list, node)
			continue
		}

		// Perform a list replacement modification
		if modificationDefinition.listOp == listOpReplace {
			result = make([]interface{}, len(modificationDefinition.list))
			copy(result, modificationDefinition.list)
			continue
		}

		// Perform insert, delete, append, prepend operation
		if modificationDefinition.key == "" && modificationDefinition.name == "" { // Index comes directly from operation definition
			idx = modificationDefinition.index

			// Replace the -1 marker with the actual 'end' index of the array
			if idx == -1 {
				idx = len(result)
			}

		} else if modificationDefinition.key == "" && modificationDefinition.name != "" {
			name := modificationDefinition.name
			delete := modificationDefinition.listOp == listOpDelete
			if delete {
				// Sanity check for delete operation, ensure no orphan entries follow the operator definition
				if len(modificationDefinition.list) > 0 {
					m.Errors.Append(ansi.Errorf("@m{%s}: @R{item in array directly after} @c{(( delete \"%s\" ))} @r{must be one of the array operators 'append', 'prepend', 'delete', or 'insert'}", node, name))
					return nil
				}

				// Look up the index of the specified insertion point (based on solely on its name)
				idx = getIndexOfSimpleEntry(result, name)
				if idx < 0 {
					m.Errors.Append(ansi.Errorf("@m{%s}: @R{unable to find specified modification point with} @c{'%s'}", node, name))
					return nil
				}
			}

		} else { // Index look-up based on key and name
			key := modificationDefinition.key
			name := modificationDefinition.name
			delete := modificationDefinition.listOp == listOpDelete

			// Sanity check original list, list must contain key/id based entries
			if err := canKeyMergeArray("original", result, node, key); err != nil {
				m.Errors.Append(err)
				return nil
			}

			// Sanity check new list, depending on the operation type (delete or insert)
			if delete == false {

				// Sanity check new list, list must contain key/id based entries
				if err := canKeyMergeArray("new", modificationDefinition.list, node, key); err != nil {
					m.Errors.Append(err)
					return nil
				}

				// Since we have a way to identify indiviual entries based on their key/id, we can sanity check for possible duplicates
				for _, entry := range modificationDefinition.list {
					obj := entry.(map[interface{}]interface{})
					entryName := obj[key].(string)
					if getIndexOfEntry(result, key, entryName) > 0 {
						m.Errors.Append(ansi.Errorf("@m{%s}: @R{unable to insert, because new list entry} @c{'%s: %s'} @R{is detected multiple times}", node, key, entryName))
						return nil
					}
				}
			} else {
				// Sanity check for delete operation, ensure no orphan entries follow the operator definition
				if len(modificationDefinition.list) > 0 {
					m.Errors.Append(ansi.Errorf("@m{%s}: @R{item in array directly after} @c{(( delete %s \"%s\" ))} @r{must be one of the array operators 'append', 'prepend', 'delete', or 'insert'}", node, key, name))
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
		if modificationDefinition.relative == "after" {
			idx++
		}

		// Back out if idx is smaller than 0, or greater than the length (for inserts), or greater/equal than the length (for deletes)
		if (idx < 0) || (modificationDefinition.listOp != listOpDelete && idx > len(result)) || (modificationDefinition.listOp == listOpDelete && idx >= len(result)) {
			m.Errors.Append(ansi.Errorf("@m{%s}: @R{unable to modify the list, because specified index} @c{%d} @R{is out of bounds}", node, idx))
			return nil
		}

		if modificationDefinition.listOp != listOpDelete {
			DEBUG("%s: inserting %d new elements to existing array at index %d", node, len(modificationDefinition.list), idx)
			result = insertIntoList(result, idx, modificationDefinition.list)
		} else {
			DEBUG("%s: deleting element at array index %d", node, idx)
			result = deleteIndexFromList(result, idx)
		}
	}

	return result
}

// The magic which chooses to merge, append, or inline based on the contents of
// the array
func (m *Merger) mergeArrayDefault(orig []interface{}, n []interface{}, node string) []interface{} {
	DEBUG("%s: performing index-based array merge", node)
	var err error
	key := getDefaultIdentifierKey()

	if err = canKeyMergeArray("original", orig, node, key); err == nil {
		if err = canKeyMergeArray("new", n, node, key); err == nil {
			return m.mergeArrayByKey(orig, n, node, key)
		}
	}

	//Warn the user about any unintuitive behavior that may have gotten us here.
	if warning, isWarning := err.(WarningError); isWarning && warning.HasContext(eContextDefaultMerge) {
		mergeStratStr := "inline"
		if m.AppendByDefault {
			mergeStratStr = "append"
		}
		warning.Warn()
		NewWarningError(eContextDefaultMerge, "@Y{Falling back to %s merge strategy}", mergeStratStr).Warn()
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
		path := fmt.Sprintf("%s.%s", node, obj[key])
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

// getArrayModifications returns a list of ModificationDefinition objects with
// information on which array operations to apply to which entries. The first
// object in the returned will always represent the default merge behavior.
func getArrayModifications(obj []interface{}, simpleList bool) []ModificationDefinition {
	// Starts with an entry representing the default merge behavior
	result := []ModificationDefinition{ModificationDefinition{listOp: listOpMergeDefault}}

	// easy shortcircuit
	if len(obj) == 0 {
		return result
	}

	mergeRegEx := regexp.MustCompile("^\\Q((\\E\\s*merge\\s*\\Q))\\E$")
	mergeOnKeyRegEx := regexp.MustCompile("^\\Q((\\E\\s*merge\\s+(on)\\s+(.+)\\s*\\Q))\\E$")
	replaceRegEx := regexp.MustCompile("^\\Q((\\E\\s*replace\\s*\\Q))\\E$")
	inlineRegEx := regexp.MustCompile("^\\Q((\\E\\s*inline\\s*\\Q))\\E$")
	appendRegEx := regexp.MustCompile("^\\Q((\\E\\s*append\\s*\\Q))\\E$")
	prependRegEx := regexp.MustCompile("^\\Q((\\E\\s*prepend\\s*\\Q))\\E$")
	insertByIdxRegEx := regexp.MustCompile("^\\Q((\\E\\s*insert\\s+(after|before)\\s+(\\d+)\\s*\\Q))\\E$")
	insertByNameRegEx := regexp.MustCompile("^\\Q((\\E\\s*insert\\s+(after|before)\\s+([^ ]+)?\\s*\"(.+)\"\\s*\\Q))\\E$")
	deleteByIdxRegEx := regexp.MustCompile("^\\Q((\\E\\s*delete\\s+(-?\\d+)\\s*\\Q))\\E$")
	deleteByNameRegEx := regexp.MustCompile("^\\Q((\\E\\s*delete\\s+([^ ]+)?\\s*\"(.+)\"\\s*\\Q))\\E$")
	deleteByNameUnquotedRegEx := regexp.MustCompile("^\\Q((\\E\\s*delete\\s+([^ ]+)?\\s*(.+)\\s*\\Q))\\E$")

	for _, entry := range obj {
		e, isString := entry.(string)
		switch {
		case !isString:
			//Do absolutely nothing

		case mergeRegEx.MatchString(e): // check for (( merge ))
			result = append(result, ModificationDefinition{listOp: listOpMergeOnKey})
			continue

		case mergeOnKeyRegEx.MatchString(e): // check for (( merge on "key" ))
			/* #0 is the whole string,
			 * #1 is string 'on'
			 * #2 is the named-entry identifying key
			 */
			if captures := mergeOnKeyRegEx.FindStringSubmatch(e); len(captures) == 3 {
				key := strings.TrimSpace(captures[2])
				result = append(result, ModificationDefinition{listOp: listOpMergeOnKey, key: key})
				continue
			}

		case inlineRegEx.MatchString(e): // check for (( inline ))
			result = append(result, ModificationDefinition{listOp: listOpMergeInline})
			continue

		case replaceRegEx.MatchString(e): // check for (( replace ))
			result = append(result, ModificationDefinition{listOp: listOpReplace})
			continue

		case appendRegEx.MatchString(e): // check for (( append ))
			result = append(result, ModificationDefinition{listOp: listOpInsert, index: -1})
			continue

		case prependRegEx.MatchString(e): // check for (( prepend ))
			result = append(result, ModificationDefinition{listOp: listOpInsert, index: 0})
			continue

		case insertByIdxRegEx.MatchString(e): // check for (( insert ... <idx> ))
			/* #0 is the whole string,
			 * #1 is after or before
			 * #2 is the insertion index
			 */
			if captures := insertByIdxRegEx.FindStringSubmatch(e); len(captures) == 3 {
				relative := strings.TrimSpace(captures[1])
				position := strings.TrimSpace(captures[2])
				if idx, err := strconv.Atoi(position); err == nil {
					result = append(result, ModificationDefinition{listOp: listOpInsert, index: idx, relative: relative})
					continue
				}
			}

		case insertByNameRegEx.MatchString(e): // check for (( insert ... "<name>" ))
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
					key = getDefaultIdentifierKey()
				}

				result = append(result, ModificationDefinition{listOp: listOpInsert, relative: relative, key: key, name: name})
				continue
			}

		case deleteByIdxRegEx.MatchString(e): // check for (( delete <idx> ))
			/* #0 is the whole string,
			 * #1 is idx
			 */
			if captures := deleteByIdxRegEx.FindStringSubmatch(e); len(captures) == 2 {
				position := strings.TrimSpace(captures[1])
				if idx, err := strconv.Atoi(position); err == nil {
					result = append(result, ModificationDefinition{listOp: listOpDelete, index: idx})
					continue
				}
			}

		case deleteByNameRegEx.MatchString(e): // check for (( delete "<name>" ))
			/* #0 is the whole string,
			 * #1 contains the optional '<key>' string
			 * #2 is finally the target "<name>" string
			 */
			if captures := deleteByNameRegEx.FindStringSubmatch(e); len(captures) == 3 {
				key := strings.TrimSpace(captures[1])
				name := strings.TrimSpace(captures[2])

				// illegal state for simple lists, if you have a text with whitespaces, we want to enforce people using quotes
				if simpleList && key != "" {
					continue
				}

				if !simpleList && key == "" {
					key = getDefaultIdentifierKey()
				}

				result = append(result, ModificationDefinition{listOp: listOpDelete, key: key, name: name})
				continue
			}

		case deleteByNameUnquotedRegEx.MatchString(e): // check for (( delete "<name>" ))
			/* #0 is the whole string,
			 * #1 contains the optional '<key>' string
			 * #2 is finally the target "<name>" string
			 */
			if captures := deleteByNameUnquotedRegEx.FindStringSubmatch(e); len(captures) == 3 {
				key := strings.TrimSpace(captures[1])
				name := strings.TrimSpace(captures[2])

				// illegal state for simple lists, if you have a text with whitespaces, we want to enforce people using quotes
				if simpleList && key != "" && name != "" {
					continue
				}

				if name == "" {
					name = key
					if !simpleList {
						key = getDefaultIdentifierKey()
					} else {
						key = ""
					}
				}

				result = append(result, ModificationDefinition{listOp: listOpDelete, key: key, name: name})
				continue
			}
		}

		lastResultIdx := len(result) - 1

		// Add the current entry to the 'current' modification definition record (gathering the list)
		result[lastResultIdx].list = append(result[lastResultIdx].list, entry)
	}

	return result
}

func isSimpleList(list []interface{}) bool {
	DEBUG("Going to validate if this is a simple list: %v", list)

	if len(list) == 0 {
		return false
	}

	var hash_count int
	for _, item := range list {
		switch item.(type) {
		case map[interface{}]interface{}:
			hash_count = hash_count + 1
		}
	}
	if hash_count == 0 {
		DEBUG("Working on a simple list")
	}
	return hash_count == 0
}

func shouldKeyMergeArray(obj []interface{}) (bool, string) {
	key := getDefaultIdentifierKey()

	if len(obj) >= 1 && obj[0] != nil && reflect.TypeOf(obj[0]).Kind() == reflect.String {
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
		if o == nil {
			return ansi.Errorf("@m{%s.%d}: @R{%s object is nil - cannot merge by key}", node, i, disp)
		}
		if reflect.TypeOf(o).Kind() != reflect.Map {
			return ansi.Errorf("@m{%s.%d}: @R{%s object is a} @c{%s}@R{, not a} @c{map} @R{- cannot merge by key}", node, i, disp, reflect.TypeOf(o).Kind().String())
		}

		obj := o.(map[interface{}]interface{})
		if _, ok := obj[key]; !ok {
			return ansi.Errorf("@m{%s.%d}: @R{%s object does not contain the key} @c{'%s'}@R{ - cannot merge by key}", node, i, disp, key)
		}

		//Verify that the target key has a hashable value (i.e. a value that is not itself a hash or sequence)
		targetValue := obj[key]
		_, isMap := targetValue.(map[interface{}]interface{})
		_, isSlice := targetValue.([]interface{})
		if isMap || isSlice {
			return NewWarningError(eContextDefaultMerge, ansi.Sprintf("@m{%s.%d}: @R{%s object's key} @c{'%s'} @R{cannot have a value which is a hash or sequence - cannot merge by key}", node, i, disp, key))
		}
	}
	return nil
}

func getIndexOfSimpleEntry(list []interface{}, name string) int {
	for i, entry := range list {
		switch entry.(type) {
		case string:
			if entry == name {
				return i
			}
		}
	}
	return -1
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
