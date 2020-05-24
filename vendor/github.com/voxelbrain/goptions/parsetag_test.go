package goptions

import (
	"reflect"
	"testing"
)

func TestParseTag_Minimal(t *testing.T) {
	var tag string
	tag = `--name, -n, description='Some name'`
	f, e := parseStructField(reflect.ValueOf(string("")), tag)
	if e != nil {
		t.Fatalf("Tag parsing failed: %s", e)
	}
	expected := &Flag{
		Long:        "name",
		Short:       "n",
		Description: "Some name",
	}
	if !flagequal(f, expected) {
		t.Fatalf("Expected %#v, got %#v", expected, f)
	}
}

func TestParseTag_More(t *testing.T) {
	var tag string
	tag = `--name, -n, description='Some name', mutexgroup='selector', obligatory`
	f, e := parseStructField(reflect.ValueOf(string("")), tag)
	if e != nil {
		t.Fatalf("Tag parsing failed: %s", e)
	}
	expected := &Flag{
		Long:        "name",
		Short:       "n",
		Description: "Some name",
		MutexGroups: []string{"selector"},
		Obligatory:  true,
	}
	if !flagequal(f, expected) {
		t.Fatalf("Expected %#v, got %#v", expected, f)
	}
}

func TestParseTag_MultipleFlags(t *testing.T) {
	var tag string
	var e error
	tag = `--name1, --name2`
	_, e = parseStructField(reflect.ValueOf(string("")), tag)
	if e == nil {
		t.Fatalf("Parsing should have failed")
	}

	tag = `-n, -v`
	_, e = parseStructField(reflect.ValueOf(string("")), tag)
	if e == nil {
		t.Fatalf("Parsing should have failed")
	}
}

func flagequal(f1, f2 *Flag) bool {
	return f1.Short == f2.Short &&
		f1.Long == f2.Long &&
		reflect.DeepEqual(f1.MutexGroups, f2.MutexGroups) &&
		f1.Description == f2.Description &&
		f1.Obligatory == f2.Obligatory &&
		f1.WasSpecified == f2.WasSpecified
}
