package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTokenizer(t *testing.T) {
	Convey("splitQuoted()", t, func() {
		Convey("handles the empty string", func() {
			l := splitQuoted("")
			So(l, ShouldResemble, []string{})
		})

		Convey("handles a single unquoted token", func() {
			l := splitQuoted("testing.stuff")
			So(l, ShouldResemble, []string{"testing.stuff"})
		})

		Convey("handles leading and trailing whitespace", func() {
			l := splitQuoted("\ttesting.stuff    ")
			So(l, ShouldResemble, []string{"testing.stuff"})
		})

		Convey("handles multiple unquoted tokens", func() {
			l := splitQuoted("testing even.more stuff")
			So(l, ShouldResemble, []string{"testing", "even.more", "stuff"})
		})

		Convey("handles quoted tokens", func() {
			l := splitQuoted(`testing "even more" stuff`)
			So(l, ShouldResemble, []string{"testing", `"even more"`, "stuff"})
		})

		Convey("handles escaping of whitespace tokens", func() {
			l := splitQuoted(`testing even\ more stuff`)
			So(l, ShouldResemble, []string{"testing", "even more", "stuff"})
		})

		Convey("handles escaping of quote characters", func() {
			l := splitQuoted(`testing "even \"mo\" more" stuff`)
			So(l, ShouldResemble, []string{"testing", `"even "mo" more"`, "stuff"})
		})
	})
}
