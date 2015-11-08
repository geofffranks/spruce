package internal

import (
	"reflect"
	"regexp"
)

var (
	grabRegexp         = regexp.MustCompile(`^\Q((\E\s*grab\s+(.+)\s*\Q))\E$`)
	grabIfExistsRegexp = regexp.MustCompile(`^\Q((\E\s*grab_if_exists\s+(.+)\s*\Q))\E$`)
)

// ParseGrabOp - determine if an object is a (( grab ... )) call
func ParseGrabOp(o interface{}) (bool, string) {
	return parseOpFromRegexp(o, grabRegexp)
}

// ParseGrabIfExistsOp - determine if an object is a (( grab_if_exists ... )) call
func ParseGrabIfExistsOp(o interface{}) (bool, string) {
	return parseOpFromRegexp(o, grabIfExistsRegexp)
}

func parseOpFromRegexp(o interface{}, re *regexp.Regexp) (bool, string) {
	if o != nil && reflect.TypeOf(o).Kind() == reflect.String {
		if re.MatchString(o.(string)) {
			keys := re.FindStringSubmatch(o.(string))
			return true, keys[1]
		}
	}
	return false, ""
}
