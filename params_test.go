package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParamCheckerPostProcess(t *testing.T) {
	Convey("paramChecker.PostProces()", t, func() {
		over := ParamChecker{}
		Convey("returns nil, \"ignore\", nil", func() {
			Convey("when given anything other than a string", func() {
				val, action, err := over.PostProcess(12345, "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
			Convey("when given a non-'(( param .* ))' string", func() {
				val, action, err := over.PostProcess("regular old string", "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
			Convey("when given a quoted-'(( param .* ))' string", func() {
				val, action, err := over.PostProcess("\"(( param \"error message here\" ))\"", "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
		})
		Convey("Returns an error if an (( param .* )) form is found", func() {
			val, action, err := over.PostProcess(`(( param "error message here" ))`, "nodepath")
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "Missing param at nodepath: error message here")
			So(action, ShouldEqual, "error")
		})

		Convey("Handles unquoted error messages as well", func() {
			val, action, err := over.PostProcess(`(( param error message here ))`, "nodepath")
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "Missing param at nodepath: error message here")
			So(action, ShouldEqual, "error")
		})
	})
}
