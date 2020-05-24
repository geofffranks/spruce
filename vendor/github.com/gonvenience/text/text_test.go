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

package text_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gonvenience/bunt"
	. "github.com/gonvenience/text"
)

var _ = Describe("Generate random strings with fixed length", func() {
	Context("Random string with no prefix", func() {
		It("should generate a random string with fixed length", func() {
			Expect(len(RandomString(32))).To(BeEquivalentTo(32))
		})

		It("should fail when negative length is given", func() {
			defer func() {
				Expect(recover()).To(HaveOccurred())
			}()

			RandomString(-1)
		})
	})

	Context("Random string with given prefix", func() {
		It("should generate a random string with fixed length", func() {
			Expect(len(RandomStringWithPrefix("foobar", 32))).To(BeEquivalentTo(32))
		})

		It("should fail when the prefix is already longer than the fixed length", func() {
			defer func() {
				Expect(recover()).To(HaveOccurred())
			}()

			RandomStringWithPrefix("foobar", 4)
		})

		It("should fail when negative length is given", func() {
			defer func() {
				Expect(recover()).To(HaveOccurred())
			}()

			RandomStringWithPrefix("foobar", -1)
		})
	})

	Context("Text with given fixed length", func() {
		It("should create a string with the text and enough padding to fill it up to the required length", func() {
			Expect(FixedLength("Foobar", 10)).To(BeEquivalentTo("Foobar    "))
		})

		It("should trim the text if the text alone exceeds the provided desired length", func() {
			Expect(FixedLength("This text is too long", 10)).To(BeEquivalentTo("This [...]"))
		})

		It("should return the text as-is if it already has the perfect length", func() {
			Expect(FixedLength("Foobar", 6)).To(BeEquivalentTo("Foobar"))
		})

		It("should work with text containing ANSI sequences", func() {
			// "This text is too long" 21 characters
			// "This text is [...]" 18 characters
			actual := FixedLength(Sprintf("*This* text is too long"), 18)
			expected := Sprintf("*This* text is [...]")

			Expect(fmt.Sprintf("%#v", actual)).To(BeEquivalentTo(fmt.Sprintf("%#v", expected)))
		})

		It("should allow for a custom ellipsis", func() {
			Expect(FixedLength("This text is too long", 8, Sprintf("DimGray{...}"))).
				To(BeEquivalentTo(Sprintf("This DimGray{...}")))
		})
	})

	Context("Creating proper plurals", func() {
		It("should return human readable plurals", func() {
			Expect(Plural(0, "foobar")).To(BeEquivalentTo("no foobars"))
			Expect(Plural(1, "foobar")).To(BeEquivalentTo("one foobar"))
			Expect(Plural(2, "foobar")).To(BeEquivalentTo("two foobars"))
			Expect(Plural(3, "foobar")).To(BeEquivalentTo("three foobars"))
			Expect(Plural(4, "foobar")).To(BeEquivalentTo("four foobars"))
			Expect(Plural(5, "foobar")).To(BeEquivalentTo("five foobars"))
			Expect(Plural(6, "foobar")).To(BeEquivalentTo("six foobars"))
			Expect(Plural(7, "foobar")).To(BeEquivalentTo("seven foobars"))
			Expect(Plural(8, "foobar")).To(BeEquivalentTo("eight foobars"))
			Expect(Plural(9, "foobar")).To(BeEquivalentTo("nine foobars"))
			Expect(Plural(10, "foobar")).To(BeEquivalentTo("ten foobars"))
			Expect(Plural(11, "foobar")).To(BeEquivalentTo("eleven foobars"))
			Expect(Plural(12, "foobar")).To(BeEquivalentTo("twelve foobars"))
			Expect(Plural(13, "foobar")).To(BeEquivalentTo("13 foobars"))
			Expect(Plural(147, "foobar")).To(BeEquivalentTo("147 foobars"))

			Expect(Plural(1, "basis", "bases")).To(BeEquivalentTo("one basis"))
			Expect(Plural(2, "basis", "bases")).To(BeEquivalentTo("two bases"))
		})
	})

	Context("Creating human readable lists", func() {
		It("should create a human readable list of no strings", func() {
			Expect(List([]string{})).
				To(BeEquivalentTo(""))
		})

		It("should create a human readable list of one strings", func() {
			Expect(List([]string{"one"})).
				To(BeEquivalentTo("one"))
		})

		It("should create a human readable list of two strings", func() {
			Expect(List([]string{"one", "two"})).
				To(BeEquivalentTo("one and two"))
		})

		It("should create a human readable list of multiple strings", func() {
			Expect(List([]string{"one", "two", "three", "four"})).
				To(BeEquivalentTo("one, two, three, and four"))
		})
	})
})
