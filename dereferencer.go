package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// DeReferencer is an implementation of PostProcessor to de-reference (( grab me.data )) calls
type DeReferencer struct {
	root map[interface{}]interface{}
	ttl  int
}

// parseGrabOp - determine if an object is a (( grab ... )) call
func parseGrabOp(o interface{}) (bool, string) {
	if o != nil && reflect.TypeOf(o).Kind() == reflect.String {
		re := regexp.MustCompile(`^\Q((\E\s*grab\s+(.+)\s*\Q))\E$`)
		if re.MatchString(o.(string)) {
			keys := re.FindStringSubmatch(o.(string))
			return true, keys[1]
		}
	}
	return false, ""
}

// resolve - resolves a set of tokens (literals or references), co-recursively with resolveKey()
func (d DeReferencer) resolve(node string, args string) (interface{}, error) {
	DEBUG("%s: resolving (( grab %s )))", node, args)
	re := regexp.MustCompile(`\s+`)
	targets := re.Split(strings.Trim(args, " \t\r\n"), -1)

	if len(targets) <= 1 {
		val, err := d.resolveKey(targets[0])
		return val, err
	}
	val := []interface{}{}
	for _, target := range targets {
		v, err := d.resolveKey(target)
		if err != nil {
			return nil, err
		}
		if v != nil && reflect.TypeOf(v).Kind() == reflect.Slice {
			for i := 0; i < reflect.ValueOf(v).Len(); i++ {
				val = append(val, reflect.ValueOf(v).Index(i).Interface())
			}
		} else {
			val = append(val, v)
		}
	}
	return val, nil
}

// resolveKey - resolves a single key reference, co-recursively with resolve()
func (d DeReferencer) resolveKey(key string) (interface{}, error) {
	DEBUG("  -> resolving reference to `%s`", key)
	val, err := resolveNode(key, d.root)
	if err != nil {
		return nil, fmt.Errorf("Unable to resolve `%s`: `%s", key, err)
	}

	if should, args := parseGrabOp(val); should {
		if d.ttl -= 1; d.ttl <= 0 {
			return "", fmt.Errorf("possible recursion detected in call to (( grab ))")
		}
		val, err = d.resolve(key, args)
		d.ttl += 1
		return val, err
	}
	return val, nil
}

// PostProcess - resolves a value by seeing if it matches (( grab me.data )) and retrieves me.data's value
func (d DeReferencer) PostProcess(o interface{}, node string) (interface{}, string, error) {
	if should, args := parseGrabOp(o); should {
		d.ttl = 64
		val, err := d.resolve(node, args)
		if err != nil {
			return nil, "error", fmt.Errorf("%s: %s", node, err.Error())
		}
		DEBUG("%s: setting to %#v", node, val)
		return val, "replace", nil
	}
	return nil, "ignore", nil
}
