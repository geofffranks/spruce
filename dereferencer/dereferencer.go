package dereferencer

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/geofffranks/spruce/dereferencer/internal"
	"github.com/geofffranks/spruce/resolve"
	"github.com/geofffranks/spruce/utils"
)

// dereferencer is an implementation of PostProcessor to de-reference (( grab me.data )) calls
type dereferencer struct {
	root    map[interface{}]interface{}
	bounder *internal.RecursiveCallBounder
}

func NewDereferencer(root map[interface{}]interface{}) *dereferencer {
	return &dereferencer{
		root:    root,
		bounder: new(internal.RecursiveCallBounder),
	}
}

// PostProcess - resolves a value by seeing if it matches (( grab me.data )) or (( grab_if_exists me.data )) and retrieves me.data's value
func (d *dereferencer) PostProcess(o interface{}, node string) (interface{}, string, error) {
	d.bounder.Reset()

	var val interface{}
	var err error

	if should, args := internal.ParseGrabOp(o); should {
		val, err = d.resolveGrab(node, args)
	} else if should, args := internal.ParseGrabIfExistsOp(o); should {
		val, err = d.resolveGrabIfExists(node, args)
	} else {
		return nil, "ignore", nil
	}

	if err != nil {
		return nil, "error", fmt.Errorf("%s: %s", node, err.Error())
	}
	utils.DEBUG("%s: setting to %#v", node, val)
	return val, "replace", nil
}

var spaceRegexp = regexp.MustCompile(`\s+`)

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
	utils.DEBUG("%s: resolving (( grab_if_exists %s )))", node, args)
	targets := spaceRegexp.Split(strings.Trim(args, " \t\r\n"), -1)

	if len(targets) <= 1 {
		val, err := d.resolveKey(targets[0])
		switch err {
		case nil:
			return val, nil
		case internal.ErrInfiniteRecursion:
			return nil, err
		default:
			return nil, nil
		}
	}
	val := []interface{}{}
	for _, target := range targets {
		v, err := d.resolveKey(target)
		if err == internal.ErrInfiniteRecursion {
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
	utils.DEBUG("  -> resolving reference to `%s`", key)
	val, err := resolve.ResolveNode(key, d.root)
	if err != nil {
		return nil, fmt.Errorf("Unable to resolve `%s`: `%s", key, err)
	}

	if should, args := internal.ParseGrabOp(val); should {
		return d.bounder.Call(func() (interface{}, error) {
			return d.resolveGrab(key, args)
		})
	} else if should, args := internal.ParseGrabIfExistsOp(val); should {
		return d.bounder.Call(func() (interface{}, error) {
			return d.resolveGrabIfExists(key, args)
		})
	} else {
		return val, nil
	}
}
