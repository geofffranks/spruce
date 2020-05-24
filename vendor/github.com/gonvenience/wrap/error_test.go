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

package wrap_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gonvenience/wrap"
)

var _ = Describe("wrap package tests", func() {
	var exampleErr = fmt.Errorf("failed to do x, because of y")

	Context("wrapping errors in context", func() {
		var err = Error(
			exampleErr,
			"issue setting up z",
		)

		It("should behave and render like a standard error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("issue setting up z: failed to do x, because of y"))
		})

		It("should fall back to a simple error if no cause is provided", func() {
			err := Error(nil, "failed to do thing A")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("failed to do thing A"))
		})

		It("should be able to just extract the context string", func() {
			switch contextError := err.(type) {
			case ContextError:
				Expect(contextError.Context()).To(BeEquivalentTo("issue setting up z"))

			default:
				Fail("failed to type cast to ContextError")
			}
		})

		It("should be able to just extract the error cause", func() {
			switch contextError := err.(type) {
			case ContextError:
				Expect(contextError.Cause().Error()).To(BeEquivalentTo("failed to do x, because of y"))

			default:
				Fail("failed to type cast to ContextError")
			}
		})
	})

	Context("wrapping multiple errors with context", func() {
		var (
			err = Errorsf(
				[]error{
					fmt.Errorf("issue setting up x"),
					fmt.Errorf("issue setting up y"),
					fmt.Errorf("issue setting up z"),
				},
				"failed to setup component %s", "A",
			)
		)

		It("should behave and render like a standard error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("failed to setup component A:\n- issue setting up x\n- issue setting up y\n- issue setting up z"))
		})

		It("should fall back to a simple error if no cause is provided", func() {
			err := Errors(nil, "failed to do thing A")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("failed to do thing A"))
		})

		It("should render like a simple wrapped error if there is only one error list entry", func() {
			err := Errorsf(
				[]error{
					fmt.Errorf("issue setting up x"),
				},
				"failed to setup component %s", "A",
			)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("failed to setup component A: issue setting up x"))
		})

		It("should be able to just extract the context string", func() {
			switch contextError := err.(type) {
			case ListOfErrors:
				Expect(contextError.Context()).To(BeEquivalentTo("failed to setup component A"))

			default:
				Fail("failed to type cast to ContextError")
			}
		})

		It("should be able to just extract the list of errors", func() {
			switch contextError := err.(type) {
			case ListOfErrors:
				Expect(contextError.Errors()).To(BeEquivalentTo([]error{
					fmt.Errorf("issue setting up x"),
					fmt.Errorf("issue setting up y"),
					fmt.Errorf("issue setting up z"),
				}))

			default:
				Fail("failed to type cast to ContextError")
			}
		})
	})

	Context("projects using wrap package", func() {
		It("should be possible to use an error to wrap with context", func() {
			err := Errorf(exampleErr,
				"unable to set up %s and %s",
				"A",
				"B")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(BeEquivalentTo("unable to set up A and B: failed to do x, because of y"))
		})
	})
})
