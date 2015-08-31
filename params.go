package main

import (
	"fmt"
	"reflect"
	"regexp"
)

// ParamChecker is an implementation of PostProcessor to check for (( param "error msg" )) calls
type ParamChecker struct {
}

// Action returns the Action string for the ParamChecker
func (over ParamChecker) Action() string {
	return "param"
}

// PostProcess - check if stale (( param "error msg" )) references exist, and error out
func (over ParamChecker) PostProcess(o interface{}, node string) (interface{}, string, error) {
	if reflect.TypeOf(o).Kind() == reflect.String {
		re := regexp.MustCompile(`^\Q((\E\s*param\s+"?(.+?)"?\s*\Q))\E$`)
		if re.MatchString(o.(string)) {
			keys := re.FindStringSubmatch(o.(string))
			if keys[1] != "" {
				return nil, "error", fmt.Errorf("Missing param at %s: %s", node, keys[1])
			}
		}
	}

	return nil, "ignore", nil
}
