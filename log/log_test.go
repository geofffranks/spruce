package log_test

import (
	"bytes"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/geofffranks/spruce/log"
)

var _ = Describe("Log", func() {
	var buf *bytes.Buffer

	BeforeEach(func() {
		buf = new(bytes.Buffer)
		log.Output = buf
	})

	AfterEach(func() {
		log.Output = os.Stderr
	})

	Describe("DEBUG", func() {
		Context("when DebugOn is true", func() {
			BeforeEach(func() {
				log.DebugOn = true
			})

			AfterEach(func() {
				log.DebugOn = false
			})

			It("outputs the debug message", func() {
				log.DEBUG("test debugging")
				Expect(buf.String()).To(Equal("DEBUG> test debugging\n"))
			})

			It("prefixes each line of multi-line input", func() {
				log.DEBUG("test debugging\nsecond line")
				Expect(buf.String()).To(Equal("DEBUG> test debugging\nDEBUG> second line\n"))
			})
		})

		Context("when DebugOn is false", func() {
			BeforeEach(func() {
				log.DebugOn = false
			})

			It("does not output anything", func() {
				log.DEBUG("test debugging")
				Expect(buf.String()).To(BeEmpty())
			})
		})
	})

	Describe("TRACE", func() {
		Context("when TraceOn is true", func() {
			BeforeEach(func() {
				log.TraceOn = true
			})

			AfterEach(func() {
				log.TraceOn = false
			})

			It("outputs the trace message", func() {
				log.TRACE("test tracing")
				Expect(buf.String()).To(Equal("-----> test tracing\n"))
			})

			It("prefixes each line of multi-line input", func() {
				log.TRACE("test tracing\nsecond line")
				Expect(buf.String()).To(Equal("-----> test tracing\n-----> second line\n"))
			})
		})

		Context("when TraceOn is false", func() {
			BeforeEach(func() {
				log.TraceOn = false
			})

			It("does not output anything", func() {
				log.TRACE("test tracing")
				Expect(buf.String()).To(BeEmpty())
			})
		})
	})
})
