// Copyright © 2019 The Homeport Team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package neat_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gonvenience/neat"

	"github.com/gonvenience/bunt"
	"github.com/gonvenience/wrap"
	"github.com/pkg/errors"
)

var _ = Describe("error rendering", func() {
	BeforeEach(func() {
		bunt.ColorSetting = bunt.ON
		bunt.TrueColorSetting = bunt.ON
	})

	AfterEach(func() {
		bunt.ColorSetting = bunt.OFF
		bunt.TrueColorSetting = bunt.OFF
	})

	Context("rendering errors", func() {
		It("should render a context error using a box", func() {
			context := fmt.Sprintf("unable to start %s", "Z")
			cause := fmt.Errorf("failed to load X and Y")
			err := wrap.Error(cause, context)

			Expect(SprintError(err)).To(
				BeEquivalentTo(ContentBox(
					"Error: "+context,
					cause.Error(),
					HeadlineColor(bunt.OrangeRed),
					ContentColor(bunt.Red),
				)))
		})

		It("should render to a writer", func() {
			buf := bytes.Buffer{}
			out := bufio.NewWriter(&buf)
			FprintError(out, fmt.Errorf("failed to do X"))
			out.Flush()

			Expect(buf.String()).To(
				BeEquivalentTo(ContentBox(
					"Error",
					"failed to do X",
					HeadlineColor(bunt.OrangeRed),
					ContentColor(bunt.Red),
				)))
		})

		It("should render to stdout, too", func() {
			captureStdout := func(f func()) string {
				r, w, err := os.Pipe()
				Expect(err).ToNot(HaveOccurred())

				tmp := os.Stdout
				defer func() {
					os.Stdout = tmp
				}()

				os.Stdout = w
				f()
				w.Close()

				var buf bytes.Buffer
				io.Copy(&buf, r)

				return buf.String()
			}

			Expect(captureStdout(func() {
				PrintError(fmt.Errorf("failed to do X"))
			})).To(
				BeEquivalentTo(ContentBox(
					"Error",
					"failed to do X",
					HeadlineColor(bunt.OrangeRed),
					ContentColor(bunt.Red),
				)))
		})

		It("should render a context error inside a context error", func() {
			root := fmt.Errorf("unable to load X")
			cause := wrap.Errorf(root, "failed to start Y")
			context := fmt.Sprintf("cannot process Z")

			err := wrap.Error(cause, context)
			Expect(SprintError(err)).To(
				BeEquivalentTo(ContentBox(
					"Error: "+context,
					SprintError(cause),
					HeadlineColor(bunt.OrangeRed),
					ContentColor(bunt.Red),
				)))
		})

		It("should render github.com/pkg/errors package errors", func() {
			message := fmt.Sprintf("unable to start Z")
			cause := fmt.Errorf("failed to load X and Y")
			err := errors.Wrap(cause, message)

			Expect(SprintError(err)).To(
				BeEquivalentTo(ContentBox(
					"Error: "+message,
					cause.Error(),
					HeadlineColor(bunt.OrangeRed),
					ContentColor(bunt.Red),
				)))
		})
	})
})
