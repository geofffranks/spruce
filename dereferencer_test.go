package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDereferencerPostProcess(t *testing.T) {
	Convey("dereferencer.PostProcess()", t, func() {
		deref := NewDereferencer(map[interface{}]interface{}{
			"value": map[interface{}]interface{}{
				"to": map[interface{}]interface{}{
					"find": "dereferenced value",
				},
			},
			"othervalue": map[interface{}]interface{}{
				"to": map[interface{}]interface{}{
					"find": "other value",
				},
			},
			"references": map[interface{}]interface{}{
				"other": map[interface{}]interface{}{
					"value": "(( grab othervalue.to.find ))",
				},
			},
			"recursion":   "(( grab corecursion ))",
			"corecursion": "(( grab recursion ))",
		})
		Convey("when given anything other than a string", func() {
			Convey("returns nil, \"ignore\", nil", func() {
				val, action, err := deref.PostProcess(12345, "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
		})
		Convey("when given a '(( prune ))' string", func() {
			Convey("returns nil, \"ignore\", nil", func() {
				val, action, err := deref.PostProcess("(( prune ))", "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
		})
		Convey("when given a non-'(( grab .* ))' string", func() {
			Convey("returns nil, \"ignore\", nil", func() {
				val, action, err := deref.PostProcess("regular old string", "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
		})
		Convey("when given a quoted-'(( grab .* ))' string", func() {
			Convey("returns nil, \"ignore\", nil", func() {
				val, action, err := deref.PostProcess("\"(( grab value.to.find ))\"", "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
		})
		Convey("when given a (( grab .* )) string", func() {
			Convey("Returns an error if resolveNode() had an error resolving", func() {
				val, action, err := deref.PostProcess("(( grab value.to.retrieve ))", "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldStartWith, "nodepath: Unable to resolve `value.to.retrieve`:")
				So(action, ShouldEqual, "error")
			})
			Convey("Returns value, \"replace\", nil on successful dereference", func() {
				val, action, err := deref.PostProcess("(( grab value.to.find ))", "nodepath")
				So(val, ShouldEqual, "dereferenced value")
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "replace")
			})
			Convey("Handles multiple dereference requests inline by returning an array", func() {
				val, action, err := deref.PostProcess("(( grab value.to.find othervalue.to.find ))", "nodepath")
				So(val, ShouldResemble, []interface{}{
					"dereferenced value",
					"other value",
				})
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "replace")
			})
			Convey("Handles references that grab other references", func() {
				val, action, err := deref.PostProcess("(( grab references.other.value ))", "nodepath")
				So(val, ShouldEqual, "other value")
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "replace")
			})
			Convey("Errors on first problem of a multiple reference request", func() {
				val, action, err := deref.PostProcess("(( grab value.to.find undefined.val othervalue.to.find ))", "nodepath")
				So(val, ShouldBeNil)
				So(action, ShouldEqual, "error")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "nodepath: Unable to resolve `undefined.val`: `undefined` could not be found in the YAML datastructure")
			})
			Convey("Extra whitespace is ok", func() {
				val, action, err := deref.PostProcess("((	  grab value.to.find		othervalue.to.find     ))", "nodepath")
				So(val, ShouldResemble, []interface{}{
					"dereferenced value",
					"other value",
				})
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "replace")
			})
			Convey("Avoids infinite recursion", func() {
				val, action, err := deref.PostProcess("(( grab recursion ))", "nodepath")
				So(val, ShouldBeNil)
				So(action, ShouldEqual, "error")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "nodepath: possible infinite recursion detected in dereferencing")
			})
		})
		Convey("when given a (( grab_if_exists .* )) string", func() {
			Convey("Returns nil if the reference cannot be resolved", func() {
				val, action, err := deref.PostProcess("(( grab_if_exists value.to.retrieve ))", "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "replace")
			})
			Convey("Returns value, \"replace\", nil on successful dereference", func() {
				val, action, err := deref.PostProcess("(( grab_if_exists value.to.find ))", "nodepath")
				So(val, ShouldEqual, "dereferenced value")
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "replace")
			})
			Convey("Handles multiple dereference requests inline by returning an array", func() {
				val, action, err := deref.PostProcess("(( grab_if_exists value.to.find othervalue.to.find ))", "nodepath")
				So(val, ShouldResemble, []interface{}{
					"dereferenced value",
					"other value",
				})
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "replace")
			})
			Convey("Handles references that grab other references", func() {
				val, action, err := deref.PostProcess("(( grab_if_exists references.other.value ))", "nodepath")
				So(val, ShouldEqual, "other value")
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "replace")
			})
			Convey("Replaces unresolvable references with null values, and other refernces normally, in a multiple reference request", func() {
				val, action, err := deref.PostProcess("(( grab_if_exists value.to.find undefined.val othervalue.to.find ))", "nodepath")
				So(val, ShouldResemble, []interface{}{
					"dereferenced value",
					nil,
					"other value",
				})
				So(action, ShouldEqual, "replace")
				So(err, ShouldBeNil)
			})
			Convey("Extra whitespace is ok", func() {
				val, action, err := deref.PostProcess("((	  grab_if_exists value.to.find		othervalue.to.find     ))", "nodepath")
				So(val, ShouldResemble, []interface{}{
					"dereferenced value",
					"other value",
				})
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "replace")
			})
			Convey("Avoids infinite recursion", func() {
				val, action, err := deref.PostProcess("(( grab_if_exists recursion ))", "nodepath")
				So(val, ShouldBeNil)
				So(action, ShouldEqual, "error")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "nodepath: possible infinite recursion detected in dereferencing")
			})
		})
	})
}
