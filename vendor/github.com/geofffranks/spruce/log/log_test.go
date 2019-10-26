package log

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDebug(t *testing.T) {
	var stderr string
	PrintfStdErr = func(format string, args ...interface{}) {
		stderr = fmt.Sprintf(format, args...)
	}
	Convey("debug", t, func() {
		Convey("Outputs when debug is set to true", func() {
			stderr = ""
			DebugOn = true
			DEBUG("test debugging")
			So(stderr, ShouldEqual, "DEBUG> test debugging\n")
		})
		Convey("Multi-line debug inputs are each prefixed", func() {
			stderr = ""
			DebugOn = true
			DEBUG("test debugging\nsecond line")
			So(stderr, ShouldEqual, "DEBUG> test debugging\nDEBUG> second line\n")
		})
		Convey("Doesn't output when debug is set to false", func() {
			stderr = ""
			DebugOn = false
			DEBUG("test debugging")
			So(stderr, ShouldEqual, "")
		})
	})
}
