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

package bunt_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gonvenience/bunt"
)

var _ = Describe("parse input string", func() {
	BeforeEach(func() {
		ColorSetting = ON
		TrueColorSetting = ON
	})

	AfterEach(func() {
		ColorSetting = AUTO
		TrueColorSetting = AUTO
	})

	Context("parse Select Graphic Rendition (SGR) based input", func() {
		It("should parse an input string with SGR parameters", func() {
			input := "Example: \x1b[1mbold\x1b[0m, \x1b[3mitalic\x1b[0m, \x1b[4munderline\x1b[0m, \x1b[38;2;133;247;7mforeground\x1b[0m, and \x1b[48;2;133;247;7mbackground\x1b[0m."
			result, err := ParseString(input)
			Expect(err).ToNot(HaveOccurred())
			Expect(*result).To(
				BeEquivalentTo(String([]ColoredRune{
					{'E', 0x0000000000000000},
					{'x', 0x0000000000000000},
					{'a', 0x0000000000000000},
					{'m', 0x0000000000000000},
					{'p', 0x0000000000000000},
					{'l', 0x0000000000000000},
					{'e', 0x0000000000000000},
					{':', 0x0000000000000000},
					{' ', 0x0000000000000000},
					{'b', 0x0000000000000004},
					{'o', 0x0000000000000004},
					{'l', 0x0000000000000004},
					{'d', 0x0000000000000004},
					{',', 0x0000000000000000},
					{' ', 0x0000000000000000},
					{'i', 0x0000000000000008},
					{'t', 0x0000000000000008},
					{'a', 0x0000000000000008},
					{'l', 0x0000000000000008},
					{'i', 0x0000000000000008},
					{'c', 0x0000000000000008},
					{',', 0x0000000000000000},
					{' ', 0x0000000000000000},
					{'u', 0x0000000000000010},
					{'n', 0x0000000000000010},
					{'d', 0x0000000000000010},
					{'e', 0x0000000000000010},
					{'r', 0x0000000000000010},
					{'l', 0x0000000000000010},
					{'i', 0x0000000000000010},
					{'n', 0x0000000000000010},
					{'e', 0x0000000000000010},
					{',', 0x0000000000000000},
					{' ', 0x0000000000000000},
					{'f', 0x0000000007F78501},
					{'o', 0x0000000007F78501},
					{'r', 0x0000000007F78501},
					{'e', 0x0000000007F78501},
					{'g', 0x0000000007F78501},
					{'r', 0x0000000007F78501},
					{'o', 0x0000000007F78501},
					{'u', 0x0000000007F78501},
					{'n', 0x0000000007F78501},
					{'d', 0x0000000007F78501},
					{',', 0x0000000000000000},
					{' ', 0x0000000000000000},
					{'a', 0x0000000000000000},
					{'n', 0x0000000000000000},
					{'d', 0x0000000000000000},
					{' ', 0x0000000000000000},
					{'b', 0x0007F78500000002},
					{'a', 0x0007F78500000002},
					{'c', 0x0007F78500000002},
					{'k', 0x0007F78500000002},
					{'g', 0x0007F78500000002},
					{'r', 0x0007F78500000002},
					{'o', 0x0007F78500000002},
					{'u', 0x0007F78500000002},
					{'n', 0x0007F78500000002},
					{'d', 0x0007F78500000002},
					{'.', 0x0000000000000000},
				})))
		})

		It("should bail out nicely in case of invalid 24 bit foreground color parameters", func() {
			input := "Invalid: \x1b[38;2;1;2mfoobar\x1b[0m."
			result, err := ParseString(input)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should bail out nicely in case of invalid 24 bit background color parameters", func() {
			input := "Invalid: \x1b[48;2;1;2mfoobar\x1b[0m."
			result, err := ParseString(input)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	Context("parse markdown style text annotations", func() {
		It("should parse an input string with markdown style text annotations", func() {
			input := "Example: *bold*, _italic_, ~underline~, and CornflowerBlue{foreground}."
			result, err := ParseString(input, ProcessTextAnnotations())
			Expect(err).ToNot(HaveOccurred())
			Expect(*result).To(
				BeEquivalentTo(String([]ColoredRune{
					{'E', 0x0000000000000000},
					{'x', 0x0000000000000000},
					{'a', 0x0000000000000000},
					{'m', 0x0000000000000000},
					{'p', 0x0000000000000000},
					{'l', 0x0000000000000000},
					{'e', 0x0000000000000000},
					{':', 0x0000000000000000},
					{' ', 0x0000000000000000},
					{'b', 0x0000000000000004},
					{'o', 0x0000000000000004},
					{'l', 0x0000000000000004},
					{'d', 0x0000000000000004},
					{',', 0x0000000000000000},
					{' ', 0x0000000000000000},
					{'i', 0x0000000000000008},
					{'t', 0x0000000000000008},
					{'a', 0x0000000000000008},
					{'l', 0x0000000000000008},
					{'i', 0x0000000000000008},
					{'c', 0x0000000000000008},
					{',', 0x0000000000000000},
					{' ', 0x0000000000000000},
					{'u', 0x0000000000000010},
					{'n', 0x0000000000000010},
					{'d', 0x0000000000000010},
					{'e', 0x0000000000000010},
					{'r', 0x0000000000000010},
					{'l', 0x0000000000000010},
					{'i', 0x0000000000000010},
					{'n', 0x0000000000000010},
					{'e', 0x0000000000000010},
					{',', 0x0000000000000000},
					{' ', 0x0000000000000000},
					{'a', 0x0000000000000000},
					{'n', 0x0000000000000000},
					{'d', 0x0000000000000000},
					{' ', 0x0000000000000000},
					{'f', 0x00000000ED956401},
					{'o', 0x00000000ED956401},
					{'r', 0x00000000ED956401},
					{'e', 0x00000000ED956401},
					{'g', 0x00000000ED956401},
					{'r', 0x00000000ED956401},
					{'o', 0x00000000ED956401},
					{'u', 0x00000000ED956401},
					{'n', 0x00000000ED956401},
					{'d', 0x00000000ED956401},
					{'.', 0x0000000000000000},
				})))
		})

		It("should bail out nicely in case of invalid color name", func() {
			input := "Invalid: InvalidColor{foobar}."
			result, err := ParseString(input, ProcessTextAnnotations())
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	Context("parse input string with multi-byte characters", func() {
		It("should correctly parse multi-byte character strings", func() {
			input := "Gray{Debug➤ Found asset file} White{X} with permission White{Y}"
			result, err := ParseString(input, ProcessTextAnnotations())
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.String()).To(
				BeEquivalentTo("\x1b[38;2;128;128;128mDebug➤ Found asset file\x1b[0m \x1b[38;2;255;255;255mX\x1b[0m with permission \x1b[38;2;255;255;255mY\x1b[0m"))
		})
	})
})
