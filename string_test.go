package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTokenizer(t *testing.T) {
	Convey("splitQuoted()", t, func() {
		Convey("handles the empty string", func() {
			l, err := splitQuoted("")
			So(err, ShouldBeNil)
			So(l, ShouldResemble, []string{})
		})

		Convey("handles a single unquoted token", func() {
			l, err := splitQuoted("testing.stuff")
			So(err, ShouldBeNil)
			So(l, ShouldResemble, []string{"testing.stuff"})
		})

		Convey("handles leading and trailing whitespace", func() {
			l, err := splitQuoted("\ttesting.stuff    ")
			So(err, ShouldBeNil)
			So(l, ShouldResemble, []string{"testing.stuff"})
		})

		Convey("handles multiple unquoted tokens", func() {
			l, err := splitQuoted("testing even.more stuff")
			So(err, ShouldBeNil)
			So(l, ShouldResemble, []string{"testing", "even.more", "stuff"})
		})

		Convey("handles quoted tokens", func() {
			l, err := splitQuoted(`testing "even more" stuff`)
			So(err, ShouldBeNil)
			So(l, ShouldResemble, []string{"testing", `"even more"`, "stuff"})
		})

		Convey("handles escaping of whitespace tokens", func() {
			l, err := splitQuoted(`testing even\ more stuff`)
			So(err, ShouldBeNil)
			So(l, ShouldResemble, []string{"testing", "even more", "stuff"})
		})

		Convey("handles escaping of quote characters", func() {
			l, err := splitQuoted(`testing "even \"mo\" more" stuff`)
			So(err, ShouldBeNil)
			So(l, ShouldResemble, []string{"testing", `"even "mo" more"`, "stuff"})
		})
	})
}
