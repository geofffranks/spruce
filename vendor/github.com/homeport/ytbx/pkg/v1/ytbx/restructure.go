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

package ytbx

import (
	"encoding/json"
	"sort"

	yaml "gopkg.in/yaml.v2"
)

// DisableRemainingKeySort disables that that during restructuring of map keys,
// all unknown keys are also sorted in such a way that it ideally improves the
// readability.
var DisableRemainingKeySort = false

var knownKeyOrders = [][]string{
	{"name", "director_uuid", "releases", "instance_groups", "networks", "resource_pools", "compilation"},
	{"name", "url", "version", "sha1"},

	// Concourse (https://concourse-ci.org/pipelines.html, https://concourse-ci.org/steps.html, https://concourse-ci.org/resources.html)
	{"jobs", "resources", "resource_types"},
	{"name", "type", "source"},
	{"get"},
	{"put"},
	{"task"},

	// SUSE SCF role manifest (https://github.com/SUSE/scf/blob/develop/container-host-files/etc/scf/config/role-manifest.yml)
	{"releases", "instance_groups", "configuration", "variables"},
	{"auth", "templates"},

	// Universal default #1 ... name should always be first
	{"name"},

	// Universal default #2 ... key should always be first
	{"key"},

	// Universal default #3 ... id should always be first
	{"id"},
}

func lookupMap(list []string) map[string]int {
	result := make(map[string]int, len(list))
	for idx, entry := range list {
		result[entry] = idx
	}

	return result
}

func countCommonKeys(keys []string, list []string) int {
	counter, lookup := 0, lookupMap(keys)
	for _, key := range list {
		if _, ok := lookup[key]; ok {
			counter++
		}
	}

	return counter
}

func commonKeys(setA []string, setB []string) []string {
	result, lookup := []string{}, lookupMap(setB)
	for _, entry := range setA {
		if _, ok := lookup[entry]; ok {
			result = append(result, entry)
		}
	}

	return result
}

func reorderMapsliceKeys(input yaml.MapSlice, keys []string) yaml.MapSlice {
	// Create list with all remaining keys: those that are part of the input
	// YAML MapSlice, but not listed in the keys list
	remainingKeys, lookup := []string{}, lookupMap(keys)
	for _, mapitem := range input {
		key := mapitem.Key.(string)
		if _, ok := lookup[key]; !ok {
			remainingKeys = append(remainingKeys, key)
		}
	}

	// Sort remaining keys by sorting long and possibly hard to read structure
	// to the end of the map
	if !DisableRemainingKeySort {
		sort.Slice(remainingKeys, func(i, j int) bool {
			valI, _ := getValueByKey(input, remainingKeys[i])
			valJ, _ := getValueByKey(input, remainingKeys[j])
			marI, _ := json.Marshal(valI)
			marJ, _ := json.Marshal(valJ)
			return len(marI) < len(marJ)
		})
	}

	// Rebuild a new YAML MapSlice key by key by using first the keys from the
	// reorder list and then all remaining keys
	result := yaml.MapSlice{}
	for _, key := range append(keys, remainingKeys...) {
		// Ignore the error field here since we know what keys there are
		value, _ := getValueByKey(input, key)
		result = append(result, yaml.MapItem{
			Key:   key,
			Value: value,
		})
	}

	return result
}

func getSuitableReorderFunction(keys []string) func(yaml.MapSlice) yaml.MapSlice {
	topCandidateIdx, topCandidateHits := -1, -1
	for idx, candidate := range knownKeyOrders {
		if count := countCommonKeys(keys, candidate); count > topCandidateHits {
			topCandidateIdx = idx
			topCandidateHits = count
		}
	}

	if topCandidateIdx >= 0 {
		return func(input yaml.MapSlice) yaml.MapSlice {
			return reorderMapsliceKeys(input, commonKeys(knownKeyOrders[topCandidateIdx], keys))
		}
	}

	return nil
}

// RestructureObject takes an object and traverses down any sub elements such as list entries or map values to recursively call restructure itself. On YAML MapSlices (maps), it will use a look-up mechanism to decide if the order of key in that map needs to be rearranged to meet some known established human order.
func RestructureObject(obj interface{}) interface{} {
	switch val := obj.(type) {
	case yaml.MapSlice:
		// Restructure the YAML MapSlice keys
		if keys, err := ListStringKeys(val); err == nil {
			if fn := getSuitableReorderFunction(keys); fn != nil {
				val = fn(val)
			}
		}

		// Restructure the values of the respective keys of this YAML MapSlice
		for idx := range val {
			val[idx].Value = RestructureObject(val[idx].Value)
		}

		return val

	case []interface{}:
		for i := range val {
			val[i] = RestructureObject(val[i])
		}
		return val

	case []yaml.MapSlice:
		for i := range val {
			val[i] = RestructureObject(val[i]).(yaml.MapSlice)
		}
		return val

	default:
		return obj
	}
}
