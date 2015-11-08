package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// dereferencer is an implementation of PostProcessor to de-reference (( grab me.data )) calls
type dereferencer struct {
	root map[interface{}]interface{}
	*recursiveCallBounder
}

func NewDereferencer(root map[interface{}]interface{}) *dereferencer {
	return &dereferencer{
		root:                 root,
		recursiveCallBounder: new(recursiveCallBounder),
	}
}

// resolveGrab - resolves a set of tokens (literals or references), co-recursively with resolveKey()
func (d *dereferencer) resolveGrab(node string, args string) (interface{}, error) {
	targets := spaceRegexp.Split(strings.Trim(args, " \t\r\n"), -1)

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

// resolveGrabIfExists - resolves a set of tokens (literals or references), co-recursively with resolveKey()
func (d *dereferencer) resolveGrabIfExists(node string, args string) (interface{}, error) {
	DEBUG("%s: resolving (( grab_if_exists %s )))", node, args)
	targets := spaceRegexp.Split(strings.Trim(args, " \t\r\n"), -1)

	if len(targets) <= 1 {
		val, err := d.resolveKey(targets[0])
		switch err {
		case nil:
			return val, nil
		case errInfiniteRecursion:
			return nil, err
		default:
			return nil, nil
		}
	}
	val := []interface{}{}
	for _, target := range targets {
		v, err := d.resolveKey(target)
		if err == errInfiniteRecursion {
			return nil, err
		} else if err != nil {
			val = append(val, nil)
		} else if v != nil && reflect.TypeOf(v).Kind() == reflect.Slice {
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
func (d *dereferencer) resolveKey(key string) (interface{}, error) {
	DEBUG("  -> resolving reference to `%s`", key)
	val, err := resolveNode(key, d.root)
	if err != nil {
		return nil, fmt.Errorf("Unable to resolve `%s`: `%s", key, err)
	}

	if should, args := parseGrabOp(val); should {
		return d.recursiveCallBounder.call(func() (interface{}, error) {
			return d.resolveGrab(key, args)
		})
	} else if should, args := parseGrabIfExistsOp(val); should {
		return d.recursiveCallBounder.call(func() (interface{}, error) {
			return d.resolveGrabIfExists(key, args)
		})
	} else {
		return val, nil
	}
}

// PostProcess - resolves a value by seeing if it matches (( grab me.data )) or (( grab_if_exists me.data )) and retrieves me.data's value
func (d *dereferencer) PostProcess(o interface{}, node string) (interface{}, string, error) {
	d.recursiveCallBounder.reset()
	if should, args := parseGrabOp(o); should {
		val, err := d.resolveGrab(node, args)
		if err != nil {
			return nil, "error", fmt.Errorf("%s: %s", node, err.Error())
		}
		DEBUG("%s: setting to %#v", node, val)
		return val, "replace", nil
	} else if should, args := parseGrabIfExistsOp(o); should {
		val, err := d.resolveGrabIfExists(node, args)
		if err != nil {
			return nil, "error", fmt.Errorf("%s: %s", node, err.Error())
		}
		DEBUG("%s: setting to %#v", node, val)
		return val, "replace", nil
	} else {
		return nil, "ignore", nil
	}
}

const maxRecursiveCallLimit = 64

var errInfiniteRecursion = fmt.Errorf("possible infinite recursion detected in dereferencing")

type recursiveCallBounder struct {
	ttl int
}

func (b *recursiveCallBounder) call(f func() (interface{}, error)) (interface{}, error) {
	if b.ttl -= 1; b.ttl == 0 {
		return "", errInfiniteRecursion
	}
	defer func() { b.ttl += 1 }()
	return f()
}

func (b *recursiveCallBounder) reset() {
	b.ttl = maxRecursiveCallLimit
}

var (
	grabRegexp         = regexp.MustCompile(`^\Q((\E\s*grab\s+(.+)\s*\Q))\E$`)
	grabIfExistsRegexp = regexp.MustCompile(`^\Q((\E\s*grab_if_exists\s+(.+)\s*\Q))\E$`)
	spaceRegexp        = regexp.MustCompile(`\s+`)
)

// parseGrabOp - determine if an object is a (( grab ... )) call
func parseGrabOp(o interface{}) (bool, string) {
	return parseOpFromRegexp(o, grabRegexp)
}

// parseGrabIfExistsOp - determine if an object is a (( grab_if_exists ... )) call
func parseGrabIfExistsOp(o interface{}) (bool, string) {
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
