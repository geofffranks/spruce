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
		re := regexp.MustCompile("^\\Q((\\E\\s*grab\\s+(\\S+?)\\s*\\Q))\\E$")
		if re.MatchString(o.(string)) {
			keys := re.FindStringSubmatch(o.(string))
			if keys[1] != "" {
				target := keys[1]
				DEBUG("%s: resolving from %s", node, target)
				val, err := resolveNode(target, d.root)
				if err != nil {
					return nil, "error", fmt.Errorf("%s: Unable to resolve `%s`: `%s", node, target, err.Error())
				}
				DEBUG("%s: setting to %#v", node, val)
				return val, "replace", nil
			}
		}
	}

	return nil, "ignore", nil
}
