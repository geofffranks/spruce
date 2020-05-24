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

var _ = Describe("convenience functions", func() {
	BeforeEach(func() {
		ColorSetting = ON
		TrueColorSetting = ON
	})

	AfterEach(func() {
		ColorSetting = AUTO
		TrueColorSetting = AUTO
	})

	Context("substring function", func() {
		It("should work to correctly cut a string with ANSI sequences", func() {
			input := Substring("Text: \x1b[1mThis\x1b[0m text is too _long_", 6, 22)
			expected := "\x1b[1mThis\x1b[0m text is too"
			Expect(input).To(BeEquivalentTo(expected))
		})

		It("should panic in case string parsing inside Substring fails", func() {
			Expect(func() {
				Substring("\x1b[38;2;1;2mnot ok\x1b[0m", 0, 4)
			}).To(Panic())
		})
	})

	Context("text length function", func() {
		It("should return the correct text length of a string with ANSI sequences", func() {
			Expect(PlainTextLength("\x1b[0;32mINFO \x1b[mNo dependencies found")).To(BeEquivalentTo(len("INFO No dependencies found")))
		})

		It("should return the right size when used on strings created by the bunt package", func() {
			Expect(PlainTextLength(Sprintf("*This* text is too long"))).To(BeEquivalentTo(len(Sprintf("This text is too long"))))
		})

		It("should return the correct length based on the rune count", func() {
			Expect(PlainTextLength("fünf")).To(BeEquivalentTo(4))
		})
	})

	Context("style function", func() {
		It("should apply bold parameter to a input string", func() {
			Expect(Style("text", Bold())).To(
				BeEquivalentTo("\x1b[1mtext\x1b[0m"))
		})

		It("should apply italic parameter to a input string", func() {
			Expect(Style("text", Italic())).To(
				BeEquivalentTo("\x1b[3mtext\x1b[0m"))
		})

		It("should apply a custom foreground color to a input string", func() {
			Expect(Style("text", Foreground(CornflowerBlue))).To(
				BeEquivalentTo("\x1b[38;2;100;149;237mtext\x1b[0m"))
		})

		It("should apply the bold parameter and a custom foreground color to a input string", func() {
			Expect(Style("text", Bold(), Foreground(CornflowerBlue))).To(
				BeEquivalentTo("\x1b[1;38;2;100;149;237mtext\x1b[0m"))
		})

		It("should not evaluate special text annotations by default", func() {
			Expect(Style("_text_", Foreground(YellowGreen))).To(
				BeEquivalentTo("\x1b[38;2;154;205;50m_text_\x1b[0m"))
		})

		It("should evaluate special text annotations if enabled", func() {
			Expect(Style("_text_", Foreground(YellowGreen), EnableTextAnnotations())).To(
				BeEquivalentTo("\x1b[3;38;2;154;205;50mtext\x1b[0m"))
		})

		It("should support both line by line coloring as well as full block coloring", func() {
			// By default, color the whole string including new line sequences
			Expect(Style("text\ntext", Foreground(Yellow))).To(
				BeEquivalentTo("\x1b[38;2;255;255;0mtext\ntext\x1b[0m"))

			// If EachLine is enabled before coloring, ignore new line sequences
			Expect(Style("text\ntext", EachLine(), Foreground(Yellow))).To(
				BeEquivalentTo("\x1b[38;2;255;255;0mtext\x1b[0m\n\x1b[38;2;255;255;0mtext\x1b[0m"))

			// If EachLine is enabled after coloring, it has no effect
			Expect(Style("text\ntext", Foreground(Yellow), EachLine())).To(
				BeEquivalentTo("\x1b[38;2;255;255;0mtext\ntext\x1b[0m"))
		})

		It("should support text emphasis both line by line as well as full block mode", func() {
			Expect(Style("text\ntext", Bold())).To(
				BeEquivalentTo("\x1b[1mtext\ntext\x1b[0m"))

			Expect(Style("text\ntext", Italic())).To(
				BeEquivalentTo("\x1b[3mtext\ntext\x1b[0m"))

			Expect(Style("text\ntext", Underline())).To(
				BeEquivalentTo("\x1b[4mtext\ntext\x1b[0m"))

			Expect(Style("text\ntext", EachLine(), Bold())).To(
				BeEquivalentTo("\x1b[1mtext\x1b[0m\n\x1b[1mtext\x1b[0m"))

			Expect(Style("text\ntext", EachLine(), Italic())).To(
				BeEquivalentTo("\x1b[3mtext\x1b[0m\n\x1b[3mtext\x1b[0m"))

			Expect(Style("text\ntext", EachLine(), Underline())).To(
				BeEquivalentTo("\x1b[4mtext\x1b[0m\n\x1b[4mtext\x1b[0m"))
		})

		It("should panic in case string parsing inside Style fails", func() {
			Expect(func() {
				Style("\x1b[38;2;1;2mnot ok\x1b[0m")
			}).To(Panic())
		})

		It("should correctly apply a color to a string that already contains text emphasis", func() {
			text := Sprintf("text with *bold* and _italic_.")
			Expect(Style(text, Foreground(Orange))).To(
				BeEquivalentTo("\x1b[38;2;255;165;0mtext with \x1b[1;38;2;255;165;0mbold\x1b[0;38;2;255;165;0m and \x1b[3;38;2;255;165;0mitalic\x1b[0;38;2;255;165;0m.\x1b[0m"),
			)
		})

		It("should correctly apply a color to a string that already contains coloring", func() {
			text := Sprintf("text with Green{colored parts}.")
			Expect(Style(text, Foreground(Red))).To(
				BeEquivalentTo("\x1b[38;2;255;0;0mtext with colored parts.\x1b[0m"),
			)
		})

		It("should correctly blend a color to a string that already contains coloring", func() {
			text := Sprintf("text with Lime{colored parts}.")
			Expect(Style(text, Blend(), Foreground(Red))).To(
				BeEquivalentTo("\x1b[38;2;255;0;0mtext with \x1b[38;2;144;175;18mcolored parts\x1b[38;2;255;0;0m.\x1b[0m"),
			)
		})
	})
})
