package main

import (
	"fmt"
	"reflect"
	"regexp"
)

// DeReferencer is an implementation of PostProcessor to de-reference (( grab me.data )) calls
type DeReferencer struct {
	root map[interface{}]interface{}
}

// Action returns the Action string for the Dereferencer
func (d DeReferencer) Action() string {
	return "dereference"
}

// PostProcess - resolves a value by seeing if it matches (( grab me.data )) and retrieves me.data's value
func (d DeReferencer) PostProcess(o interface{}, node string) (interface{}, string, error) {
	if reflect.TypeOf(o).Kind() == reflect.String {
		re := regexp.MustCompile("^\\Q((\\E\\s*grab\\s+(.+?)\\s*\\Q))\\E$")
		if re.MatchString(o.(string)) {
			keys := re.FindStringSubmatch(o.(string))
			if keys[1] != "" {
				wsSquasher := regexp.MustCompile("\\s+")
				targets := wsSquasher.Split(keys[1], -1)
				if len(targets) <= 1 {
					DEBUG("%s: resolving from %s", node, targets[0])
					val, err := resolveNode(targets[0], d.root)
					if err != nil {
						return nil, "error", fmt.Errorf("%s: Unable to resolve `%s`: `%s", node, targets[0], err.Error())
					}
					DEBUG("%s: setting to %#v", node, val)
					return val, "replace", nil
				}
				val := []interface{}{}
				for _, target := range targets {
					DEBUG("%s: resolving from %s", node, target)
					v, err := resolveNode(target, d.root)
					if err != nil {
						return nil, "error", fmt.Errorf("%s: Unable to resolve `%s`: `%s", node, target, err.Error())
					}
					if reflect.TypeOf(v).Kind() == reflect.Slice {
						for i := 0; i < reflect.ValueOf(v).Len(); i++ {
							val = append(val, reflect.ValueOf(v).Index(i).Interface())
						}
					} else {
						val = append(val, v)
					}
				}
				DEBUG("%s: setting to %#v", node, val)
				return val, "replace", nil
			}
		}
	}

	return nil, "ignore", nil
}
