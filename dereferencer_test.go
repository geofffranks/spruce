package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDeReferencerAction(t *testing.T) {
	Convey("dereferencer.Action() returns correct string", t, func() {
		deref := DeReferencer{root: map[interface{}]interface{}{}}
		So(deref.Action(), ShouldEqual, "dereference")
	})
}

func TestDeReferencerPostProcess(t *testing.T) {
	Convey("dereferencer.PostProces()", t, func() {
		deref := DeReferencer{root: map[interface{}]interface{}{
			"value": map[interface{}]interface{}{
				"to": map[interface{}]interface{}{
					"find": "dereferenced value",
				},
			},
		}}
		Convey("returns nil, \"ignore\", nil", func() {
			Convey("when given anything other than a string", func() {
				val, action, err := deref.PostProcess(12345, "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
			Convey("when given a '(( prune ))' string", func() {
				val, action, err := deref.PostProcess("(( prune ))", "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
			Convey("when given a non-'(( grab .* ))' string", func() {
				val, action, err := deref.PostProcess("regular old string", "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
			Convey("when given a quoted-'(( grab .* ))' string", func() {
				val, action, err := deref.PostProcess("\"(( grab value.to.find ))\"", "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
		})
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
	})
}
