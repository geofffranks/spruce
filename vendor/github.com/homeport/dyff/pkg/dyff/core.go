// Copyright © 2019 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package dyff

import (
	"fmt"
	"strings"

	"github.com/gonvenience/bunt"
	"github.com/gonvenience/text"
	"github.com/gonvenience/wrap"
	"github.com/gonvenience/ytbx"
	"github.com/mitchellh/hashstructure"
	"github.com/texttheater/golang-levenshtein/levenshtein"
	yamlv3 "gopkg.in/yaml.v3"
)

// NonStandardIdentifierGuessCountThreshold specifies how many list entries are
// needed for the guess-the-identifier function to actually consider the key
// name. Or in short, if the lists only contain two entries each, there are more
// possibilities to find unique enough keys, which might no qualify as such.
var NonStandardIdentifierGuessCountThreshold = 3

// MinorChangeThreshold specifies how many percent of the text needs to be
// changed so that it still qualifies as being a minor string change.
var MinorChangeThreshold = 0.1

// UseGoPatchPaths style paths instead of Spruce Dot-Style
var UseGoPatchPaths = false

// CompareInputFiles is one of the convenience main entry points for comparing
// objects. In this case the representation of an input file, which might
// contain multiple documents. It returns a report with the list of differences.
func CompareInputFiles(from ytbx.InputFile, to ytbx.InputFile) (Report, error) {
	if len(from.Documents) != len(to.Documents) {
		return Report{}, fmt.Errorf("comparing YAMLs with a different number of documents is currently not supported")
	}

	result := make([]Diff, 0)
	for idx := range from.Documents {
		diffs, err := compareObjects(
			ytbx.Path{DocumentIdx: idx},
			from.Documents[idx],
			to.Documents[idx],
		)

		if err != nil {
			return Report{}, err
		}

		result = append(result, diffs...)
	}

	return Report{from, to, result}, nil
}

func compareObjects(path ytbx.Path, from *yamlv3.Node, to *yamlv3.Node) ([]Diff, error) {
	switch {
	case from == nil && to == nil:
		return []Diff{}, nil

	case (from == nil && to != nil) || (from != nil && to == nil):
		return []Diff{{
			path,
			[]Detail{{
				Kind: MODIFICATION,
				From: from,
				To:   to,
			}},
		}}, nil

	case (from.Kind != to.Kind) || (from.Tag != to.Tag):
		return []Diff{{
			path,
			[]Detail{{
				Kind: MODIFICATION,
				From: from,
				To:   to,
			}},
		}}, nil
	}

	return compareNonNilSameKindNodes(path, from, to)
}

func compareNonNilSameKindNodes(path ytbx.Path, from *yamlv3.Node, to *yamlv3.Node) ([]Diff, error) {
	var diffs []Diff
	var err error

	switch from.Kind {
	case yamlv3.DocumentNode:
		diffs, err = compareObjects(path, from.Content[0], to.Content[0])

	case yamlv3.MappingNode:
		diffs, err = compareMappingNodes(path, from, to)

	case yamlv3.SequenceNode:
		diffs, err = compareSequenceNodes(path, from, to)

	case yamlv3.ScalarNode:
		switch from.Tag {
		case "!!str":
			diffs, err = compareNodeValues(path, from, to)

		case "!!null":
			// Ignore different ways to define a null value

		default:
			if from.Value != to.Value {
				diffs, err = []Diff{{
					path,
					[]Detail{{
						Kind: MODIFICATION,
						From: from,
						To:   to,
					}},
				}}, nil
			}
		}

	case yamlv3.AliasNode:
		diffs, err = compareObjects(path, from.Alias, to.Alias)

	default:
		err = fmt.Errorf("failed to compare objects due to unsupported kind %v", from.Kind)
	}

	return diffs, err
}

func compareMappingNodes(path ytbx.Path, from *yamlv3.Node, to *yamlv3.Node) ([]Diff, error) {
	result := make([]Diff, 0)
	removals := []*yamlv3.Node{}
	additions := []*yamlv3.Node{}

	for i := 0; i < len(from.Content); i += 2 {
		key, fromItem := from.Content[i], from.Content[i+1]
		if toItem, ok := findValueByKey(to, key.Value); ok {
			// `from` and `to` contain the same `key` -> require comparison
			diffs, err := compareObjects(
				ytbx.NewPathWithNamedElement(path, key.Value),
				followAlias(fromItem),
				followAlias(toItem),
			)

			if err != nil {
				return nil, err
			}

			result = append(result, diffs...)

		} else {
			// `from` contain the `key`, but `to` does not -> removal
			removals = append(removals, key, fromItem)
		}
	}

	for i := 0; i < len(to.Content); i += 2 {
		key, toItem := to.Content[i], to.Content[i+1]
		if _, ok := findValueByKey(from, key.Value); !ok {
			// `to` contains a `key` that `from` does not have -> addition
			additions = append(additions, key, toItem)
		}
	}

	diff := Diff{Path: path, Details: []Detail{}}

	if len(removals) > 0 {
		diff.Details = append(diff.Details,
			Detail{
				Kind: REMOVAL,
				From: &yamlv3.Node{
					Kind:    from.Kind,
					Tag:     from.Tag,
					Content: removals,
				},
				To: nil,
			},
		)
	}

	if len(additions) > 0 {
		diff.Details = append(diff.Details,
			Detail{
				Kind: ADDITION,
				From: nil,
				To: &yamlv3.Node{
					Kind:    to.Kind,
					Tag:     to.Tag,
					Content: additions,
				},
			},
		)
	}

	if len(diff.Details) > 0 {
		result = append([]Diff{diff}, result...)
	}

	return result, nil
}

func compareSequenceNodes(path ytbx.Path, from *yamlv3.Node, to *yamlv3.Node) ([]Diff, error) {
	// Bail out quickly if there is nothing to check
	if len(from.Content) == 0 && len(to.Content) == 0 {
		return []Diff{}, nil
	}

	if identifier, err := getIdentifierFromNamedLists(from, to); err == nil {
		return compareNamedEntryLists(path, identifier, from, to)
	}

	if identifier := getNonStandardIdentifierFromNamedLists(from, to); identifier != "" {
		return compareNamedEntryLists(path, identifier, from, to)
	}

	return compareSimpleLists(path, from, to)
}

func compareSimpleLists(path ytbx.Path, from *yamlv3.Node, to *yamlv3.Node) ([]Diff, error) {
	removals := make([]*yamlv3.Node, 0)
	additions := make([]*yamlv3.Node, 0)

	result := make([]Diff, 0)

	fromLength := len(from.Content)
	toLength := len(to.Content)

	// Special case if both lists only contain one entry, then directly compare
	// the two entries with each other
	if fromLength == 1 && fromLength == toLength {
		return compareObjects(
			ytbx.NewPathWithIndexedListElement(path, 0),
			followAlias(from.Content[0]),
			followAlias(to.Content[0]),
		)
	}

	fromLookup := createLookUpMap(from)
	toLookup := createLookUpMap(to)

	// Fill two lists with the names of the entries that are common to both
	// provided lists
	fromNames := make([]uint64, 0, fromLength)
	toNames := make([]uint64, 0, fromLength)

	for idxPos, fromValue := range from.Content {
		hash := calcNodeHash(fromValue)

		if _, ok := toLookup[hash]; !ok {
			// `from` entry does not exist in `to` list
			removals = append(removals, from.Content[idxPos])

		} else {
			fromNames = append(fromNames, hash)
		}
	}

	for idxPos, toValue := range to.Content {
		hash := calcNodeHash(toValue)

		if _, ok := fromLookup[hash]; !ok {
			// `to` entry does not exist in `from` list
			additions = append(additions, to.Content[idxPos])

		} else {
			toNames = append(toNames, hash)
		}
	}

	return packChangesAndAddToResult(
		result,
		true,
		path,
		findOrderChangesInSimpleList(from, to, fromNames, toNames, fromLookup, toLookup),
		additions,
		removals,
	)
}

func compareNamedEntryLists(path ytbx.Path, identifier string, from *yamlv3.Node, to *yamlv3.Node) ([]Diff, error) {
	removals := make([]*yamlv3.Node, 0)
	additions := make([]*yamlv3.Node, 0)

	result := make([]Diff, 0)

	// Fill two lists with the names of the entries that are common in both lists
	fromLength := len(from.Content)
	fromNames := make([]string, 0, fromLength)
	toNames := make([]string, 0, fromLength)

	// Find entries that are common to both lists to compare them separately, and
	// find entries that are only in from, but not to and are therefore removed
	for _, fromEntry := range from.Content {
		name, err := getValueByKey(fromEntry, identifier)
		if err != nil {
			return nil, err
		}

		if toEntry, ok := getEntryFromNamedList(to, identifier, name.Value); ok {
			// `from` and `to` have the same entry idenfified by identifier and name -> require comparison
			diffs, err := compareObjects(
				ytbx.NewPathWithNamedListElement(path, identifier, name.Value),
				followAlias(fromEntry),
				followAlias(toEntry),
			)
			if err != nil {
				return nil, err
			}
			result = append(result, diffs...)
			fromNames = append(fromNames, name.Value)

		} else {
			// `from` has an entry (identified by identifier and name), but `to` does not -> removal
			removals = append(removals, fromEntry)
		}
	}

	// Find entries that are only in to, but not from and are therefore added
	for _, toEntry := range to.Content {
		name, err := getValueByKey(toEntry, identifier)
		if err != nil {
			return nil, err
		}

		if _, ok := getEntryFromNamedList(from, identifier, name.Value); ok {
			// `to` and `from` have the same entry idenfified by identifier and name (comparison already covered by previous range)
			toNames = append(toNames, name.Value)

		} else {
			// `to` has an entry (identified by identifier and name), but `from` does not -> addition
			additions = append(additions, toEntry)
		}
	}

	orderchanges := findOrderChangesInNamedEntryLists(fromNames, toNames)

	return packChangesAndAddToResult(result, true, path, orderchanges, additions, removals)
}

func compareNodeValues(path ytbx.Path, from *yamlv3.Node, to *yamlv3.Node) ([]Diff, error) {
	result := make([]Diff, 0)
	if strings.Compare(from.Value, to.Value) != 0 {
		result = append(result, Diff{
			path,
			[]Detail{{
				Kind: MODIFICATION,
				From: from,
				To:   to,
			}},
		})
	}

	return result, nil
}

func findOrderChangesInSimpleList(from, to *yamlv3.Node, fromNames, toNames []uint64, fromLookup, toLookup map[uint64]int) []Detail {
	orderchanges := make([]Detail, 0)

	cnv := func(list []uint64, lookup map[uint64]int, content *yamlv3.Node) *yamlv3.Node {
		result := make([]*yamlv3.Node, 0, len(list))
		for _, hash := range list {
			result = append(result, content.Content[lookup[hash]])
		}

		return &yamlv3.Node{
			Kind:    yamlv3.SequenceNode,
			Content: result,
		}
	}

	// Try to find order changes ...
	if len(fromNames) == len(toNames) {
		for idx, hash := range fromNames {
			if toNames[idx] != hash {
				orderchanges = append(orderchanges,
					Detail{
						Kind: ORDERCHANGE,
						From: cnv(fromNames, fromLookup, from),
						To:   cnv(toNames, toLookup, to),
					})
				break
			}
		}
	}

	return orderchanges
}

// AsSequenceNode translates a string list into a SequenceNode
func AsSequenceNode(list []string) *yamlv3.Node {
	result := make([]*yamlv3.Node, len(list))
	for i, entry := range list {
		result[i] = &yamlv3.Node{
			Kind:  yamlv3.ScalarNode,
			Tag:   "!!str",
			Value: entry,
		}
	}

	return &yamlv3.Node{
		Kind:    yamlv3.SequenceNode,
		Content: result,
	}
}

func findOrderChangesInNamedEntryLists(fromNames, toNames []string) []Detail {
	orderchanges := make([]Detail, 0)

	idxLookupMap := make(map[string]int, len(toNames))
	for idx, name := range toNames {
		idxLookupMap[name] = idx
	}

	// Try to find order changes ...
	for idx, name := range fromNames {
		if idxLookupMap[name] != idx {
			orderchanges = append(orderchanges, Detail{
				Kind: ORDERCHANGE,
				From: AsSequenceNode(fromNames),
				To:   AsSequenceNode(toNames),
			})
			break
		}
	}

	return orderchanges
}

func packChangesAndAddToResult(list []Diff, prepend bool, path ytbx.Path, orderchanges []Detail, additions, removals []*yamlv3.Node) ([]Diff, error) {
	// Prepare a diff for this path to added to the result set (if there are changes)
	diff := Diff{Path: path, Details: []Detail{}}

	if len(orderchanges) > 0 {
		diff.Details = append(diff.Details, orderchanges...)
	}

	if len(removals) > 0 {
		diff.Details = append(diff.Details, Detail{
			Kind: REMOVAL,
			From: &yamlv3.Node{
				Kind:    yamlv3.SequenceNode,
				Tag:     "!!seq",
				Content: removals,
			},
			To: nil,
		})
	}

	if len(additions) > 0 {
		diff.Details = append(diff.Details, Detail{
			Kind: ADDITION,
			From: nil,
			To: &yamlv3.Node{
				Kind:    yamlv3.SequenceNode,
				Tag:     "!!seq",
				Content: additions,
			},
		})
	}

	// If there were changes added to the details list,
	// we can safely add it to the result set.
	// Otherwise it the result set will be returned as-is.
	if len(diff.Details) > 0 {
		switch prepend {
		case true:
			list = append([]Diff{diff}, list...)

		case false:
			list = append(list, diff)
		}
	}

	return list, nil
}

func followAlias(node *yamlv3.Node) *yamlv3.Node {
	if node != nil && node.Alias != nil {
		return followAlias(node.Alias)
	}

	return node
}

func findValueByKey(mappingNode *yamlv3.Node, key string) (*yamlv3.Node, bool) {
	for i := 0; i < len(mappingNode.Content); i += 2 {
		k, v := followAlias(mappingNode.Content[i]), followAlias(mappingNode.Content[i+1])
		if k.Value == key {
			return v, true
		}
	}

	return nil, false
}

// getValueByKey returns the value for a given key in a provided mapping node,
// or nil with an error if there is no such entry. This is comparable to getting
// a value from a map with `foobar[key]`.
func getValueByKey(mappingNode *yamlv3.Node, key string) (*yamlv3.Node, error) {
	for i := 0; i < len(mappingNode.Content); i += 2 {
		k, v := followAlias(mappingNode.Content[i]), followAlias(mappingNode.Content[i+1])
		if k.Value == key {
			return v, nil
		}
	}

	if names, err := ytbx.ListStringKeys(mappingNode); err == nil {
		return nil, fmt.Errorf("no key '%s' found in map, available keys are: %s", key, strings.Join(names, ", "))
	}

	return nil, fmt.Errorf("no key '%s' found in map and also failed to get a list of key for this map", key)
}

// getEntryFromNamedList returns the entry that is identified by the identifier
// key and a name, for example: `name: one` where name is the identifier key and
// one the name. Function will return nil with bool false if there is no entry.
func getEntryFromNamedList(sequenceNode *yamlv3.Node, identifier string, name string) (*yamlv3.Node, bool) {
	for _, mappingNode := range sequenceNode.Content {
		for i := 0; i < len(mappingNode.Content); i += 2 {
			k, v := followAlias(mappingNode.Content[i]), followAlias(mappingNode.Content[i+1])
			if k.Value == identifier && v.Value == name {
				return mappingNode, true
			}
		}
	}

	return nil, false
}

func getIdentifierFromNamedLists(listA, listB *yamlv3.Node) (string, error) {
	candidates := []string{"name", "key", "id"}

	isCandidate := func(node *yamlv3.Node) bool {
		if node.Kind == yamlv3.ScalarNode {
			for _, entry := range candidates {
				if node.Value == entry {
					return true
				}
			}
		}

		return false
	}

	createKeyCountMap := func(sequenceNode *yamlv3.Node) map[string]map[string]struct{} {
		result := map[string]map[string]struct{}{}
		for _, entry := range sequenceNode.Content {
			switch entry.Kind {
			case yamlv3.MappingNode:
				for i := 0; i < len(entry.Content); i += 2 {
					k, v := followAlias(entry.Content[i]), followAlias(entry.Content[i+1])
					if isCandidate(k) {
						if _, found := result[k.Value]; !found {
							result[k.Value] = map[string]struct{}{}
						}

						result[k.Value][v.Value] = struct{}{}
					}
				}
			}
		}

		return result
	}

	counterA := createKeyCountMap(listA)
	counterB := createKeyCountMap(listB)

	// Check for the usual suspects: name, key, and id
	for _, identifier := range candidates {
		if countA, okA := counterA[identifier]; okA && len(countA) == len(listA.Content) {
			if countB, okB := counterB[identifier]; okB && len(countB) == len(listB.Content) {
				return identifier, nil
			}
		}
	}

	return "", fmt.Errorf("unable to find a key that can serve as an unique identifier")
}

func getNonStandardIdentifierFromNamedLists(listA, listB *yamlv3.Node) string {
	createKeyCountMap := func(list *yamlv3.Node) map[string]int {
		tmp := map[string]map[string]struct{}{}
		for _, entry := range list.Content {
			if entry.Kind != yamlv3.MappingNode {
				return map[string]int{}
			}

			for i := 0; i < len(entry.Content); i += 2 {
				k, v := followAlias(entry.Content[i]), followAlias(entry.Content[i+1])
				if k.Kind == yamlv3.ScalarNode && k.Tag == "!!str" &&
					v.Kind == yamlv3.ScalarNode && v.Tag == "!!str" {
					if _, ok := tmp[k.Value]; !ok {
						tmp[k.Value] = map[string]struct{}{}
					}

					tmp[k.Value][v.Value] = struct{}{}
				}
			}
		}

		result := map[string]int{}
		for key, value := range tmp {
			result[key] = len(value)
		}

		return result
	}

	listALength := len(listA.Content)
	listBLength := len(listB.Content)
	counterA := createKeyCountMap(listA)
	counterB := createKeyCountMap(listB)

	for keyA, countA := range counterA {
		if countB, ok := counterB[keyA]; ok {
			if countA == listALength && countB == listBLength && countA > NonStandardIdentifierGuessCountThreshold {
				return keyA
			}
		}
	}

	return ""
}

func createLookUpMap(sequenceNode *yamlv3.Node) map[uint64]int {
	result := make(map[uint64]int, len(sequenceNode.Content))
	for idx, entry := range sequenceNode.Content {
		result[calcNodeHash(entry)] = idx
	}

	return result
}

func basicType(node *yamlv3.Node) interface{} {
	switch node.Kind {
	case yamlv3.DocumentNode:
		panic("document nodes are not supported to be translated into a basic type")

	case yamlv3.MappingNode:
		result := map[interface{}]interface{}{}
		for i := 0; i < len(node.Content); i += 2 {
			k, v := followAlias(node.Content[i]), followAlias(node.Content[i+1])
			result[basicType(k)] = basicType(v)
		}

		return result

	case yamlv3.SequenceNode:
		result := []interface{}{}
		for _, entry := range node.Content {
			result = append(result, basicType(followAlias(entry)))
		}

		return result

	case yamlv3.ScalarNode:
		return node.Value

	case yamlv3.AliasNode:
		return basicType(node.Alias)

	default:
		panic("should be unreachable")
	}
}

func calcNodeHash(node *yamlv3.Node) uint64 {
	switch node.Kind {
	case yamlv3.MappingNode, yamlv3.SequenceNode:
		hash, err := hashstructure.Hash(basicType(node), nil)
		if err != nil {
			panic(wrap.Errorf(err, "failed to calculate hash of %#v", node))
		}

		return hash

	case yamlv3.ScalarNode:
		hash, err := hashstructure.Hash(node.Value, nil)
		if err != nil {
			panic(wrap.Errorf(err, "failed to calculate hash of %#v", node.Value))
		}

		return hash

	case yamlv3.AliasNode:
		return calcNodeHash(followAlias(node))

	default:
		panic(fmt.Errorf("failed to calculate hash of node, kind %v is not supported", node.Kind))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func isMinorChange(from string, to string) bool {
	levenshteinDistance := levenshtein.DistanceForStrings([]rune(from), []rune(to), levenshtein.DefaultOptions)

	// Special case: Consider it a minor change if only two runes/characters were
	// changed, which results in a default distance of four, two removals and two
	// additions each.
	if levenshteinDistance <= 4 {
		return true
	}

	referenceLength := min(len(from), len(to))
	return float64(levenshteinDistance)/float64(referenceLength) < MinorChangeThreshold
}

func isMultiLine(from string, to string) bool {
	return strings.Contains(from, "\n") || strings.Contains(to, "\n")
}

func isList(node *yamlv3.Node) bool {
	switch node.Kind {
	case yamlv3.SequenceNode:
		return true
	}

	return false
}

// ChangeRoot changes the root of an input file to a position inside its
// document based on the given path. Input files with more than one document are
// not supported, since they could have multiple elements with that path.
func ChangeRoot(inputFile *ytbx.InputFile, path string, translateListToDocuments bool) error {
	multipleDocuments := len(inputFile.Documents) != 1

	if multipleDocuments {
		return fmt.Errorf("change root for an input file is only possible if there is only one document, but %s contains %s",
			inputFile.Location,
			text.Plural(len(inputFile.Documents), "document"))
	}

	// For reference reasons, keep the original root level
	originalRoot := inputFile.Documents[0]

	// Find the object at the given path
	obj, err := ytbx.Grab(inputFile.Documents[0], path)
	if err != nil {
		return err
	}

	wrapInDocumentNodes := func(list []*yamlv3.Node) []*yamlv3.Node {
		result := make([]*yamlv3.Node, len(list))
		for i := range list {
			result[i] = &yamlv3.Node{
				Kind:    yamlv3.DocumentNode,
				Content: []*yamlv3.Node{list[i]},
			}
		}

		return result
	}

	if translateListToDocuments && isList(obj) {
		// Change root of input file main document to a new list of documents based on the the list that was found
		inputFile.Documents = wrapInDocumentNodes(obj.Content)

	} else {
		// Change root of input file main document to the object that was found
		inputFile.Documents = wrapInDocumentNodes([]*yamlv3.Node{obj})
	}

	// Parse path string and create nicely formatted output path
	if resolvedPath, err := ytbx.ParsePathString(path, originalRoot); err == nil {
		path = pathToString(resolvedPath, multipleDocuments)
	}

	inputFile.Note = fmt.Sprintf("YAML root was changed to %s", path)

	return nil
}

func pathToString(path ytbx.Path, showDocumentIdx bool) string {
	var result string

	if UseGoPatchPaths {
		result = styledGoPatchPath(path)

	} else {
		result = styledDotStylePath(path)
	}

	if showDocumentIdx {
		result += bunt.Sprintf("  LightSteelBlue{(document #%d)}", path.DocumentIdx+1)
	}

	return result
}
