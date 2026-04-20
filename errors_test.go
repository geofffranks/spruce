package spruce

import (
	"bytes"
	"fmt"
	"os"

	"github.com/geofffranks/spruce/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Errors", func() {
	Describe("MultiError", func() {
		It("formats multiple errors with count", func() {
			me := MultiError{Errors: []error{fmt.Errorf("first"), fmt.Errorf("second")}}
			Expect(me.Error()).To(ContainSubstring("2 error(s) detected"))
			Expect(me.Error()).To(ContainSubstring("first"))
			Expect(me.Error()).To(ContainSubstring("second"))
		})

		It("returns correct count", func() {
			me := MultiError{}
			Expect(me.Count()).To(Equal(0))
			me.Append(fmt.Errorf("one"))
			Expect(me.Count()).To(Equal(1))
		})

		Describe("Append", func() {
			It("does nothing when appending nil", func() {
				me := MultiError{}
				me.Append(nil)
				Expect(me.Count()).To(Equal(0))
			})

			It("flattens nested MultiErrors", func() {
				inner := MultiError{Errors: []error{fmt.Errorf("a"), fmt.Errorf("b")}}
				outer := MultiError{}
				outer.Append(inner)
				Expect(outer.Count()).To(Equal(2))
			})

			It("appends regular errors", func() {
				me := MultiError{}
				me.Append(fmt.Errorf("regular"))
				Expect(me.Count()).To(Equal(1))
			})
		})
	})

	Describe("WarningError", func() {
		It("returns the warning message from Error()", func() {
			we := NewWarningError(eContextAll, "test warning %s", "msg")
			Expect(we.Error()).To(ContainSubstring("test warning msg"))
		})

		It("has context for eContextAll (0)", func() {
			we := NewWarningError(eContextAll, "test")
			Expect(we.HasContext(eContextDefaultMerge)).To(BeTrue())
		})

		It("has specific context when set", func() {
			we := NewWarningError(eContextDefaultMerge, "test")
			Expect(we.HasContext(eContextDefaultMerge)).To(BeTrue())
		})

		Describe("Warn", func() {
			var buf *bytes.Buffer

			BeforeEach(func() {
				buf = new(bytes.Buffer)
				log.Output = buf
			})

			AfterEach(func() {
				log.Output = os.Stderr
			})

			It("prints warning when warnings are not silenced", func() {
				SilenceWarnings(false)
				we := NewWarningError(eContextAll, "loud warning")
				we.Warn()
				Expect(buf.String()).To(ContainSubstring("loud warning"))
			})

			It("does not print when warnings are silenced", func() {
				SilenceWarnings(true)
				defer SilenceWarnings(false)
				we := NewWarningError(eContextAll, "quiet warning")
				we.Warn()
				Expect(buf.String()).To(BeEmpty())
			})
		})
	})
})
