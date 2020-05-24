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
	"bytes"
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gonvenience/bunt"
	. "github.com/gonvenience/neat"
	. "github.com/gonvenience/term"
)

var _ = Describe("content box", func() {
	BeforeEach(func() {
		ColorSetting = ON
		TrueColorSetting = ON
	})

	AfterEach(func() {
		ColorSetting = OFF
		TrueColorSetting = OFF
	})

	Context("rendering content boxes", func() {
		var (
			headline = "headline"
			content  = `multi
line
content
`
		)

		It("should create a simply styled content box", func() {
			Expect("\n" + ContentBox(headline, content)).To(BeEquivalentTo(Sprintf(`
╭ headline
│ multi
│ line
│ content
╵
`)))
		})

		It("should create a simply styled content box with bold headline", func() {
			Expect("\n" + ContentBox(headline, content,
				HeadlineStyle(Bold()),
			)).To(BeEquivalentTo(Sprintf(`
╭ *headline*
│ multi
│ line
│ content
╵
`)))
		})

		It("should create a simply styled content box with italic, underline, and bold headline", func() {
			Expect("\n" + ContentBox(headline, content,
				HeadlineStyle(Italic()),
				HeadlineStyle(Underline()),
				HeadlineStyle(Bold()),
			)).To(BeEquivalentTo(Sprintf(`
╭ _*~headline~*_
│ multi
│ line
│ content
╵
`)))
		})

		It("should create styled content box with headline colors", func() {
			Expect("\n" + ContentBox(headline, content,
				HeadlineColor(DodgerBlue),
			)).To(BeEquivalentTo(Sprintf(`
DodgerBlue{╭} DodgerBlue{headline}
DodgerBlue{│} multi
DodgerBlue{│} line
DodgerBlue{│} content
DodgerBlue{╵}
`)))
		})

		It("should create styled content box with content colors", func() {
			Expect("\n" + ContentBox(headline, content,
				ContentColor(DimGray),
			)).To(BeEquivalentTo(Sprintf(`
╭ headline
│ DimGray{multi}
│ DimGray{line}
│ DimGray{content}
╵
`)))
		})

		It("should create styled content box with headline and content colors", func() {
			Expect("\n" + ContentBox(headline, content,
				HeadlineColor(DodgerBlue),
				ContentColor(DimGray),
			)).To(BeEquivalentTo(Sprintf(`
DodgerBlue{╭} DodgerBlue{headline}
DodgerBlue{│} DimGray{multi}
DodgerBlue{│} DimGray{line}
DodgerBlue{│} DimGray{content}
DodgerBlue{╵}
`)))
		})

		It("should create a content box with no trailing line feed", func() {
			Expect("\n" + ContentBox(
				headline,
				content,
				NoFinalEndOfLine(),
			)).To(BeEquivalentTo(Sprintf(`
╭ headline
│ multi
│ line
│ content
╵`)))
		})
	})

	Context("rendering content boxes with already colored content", func() {
		setupTestStrings := func() (string, string) {
			return Sprintf("CornflowerBlue{~headline~}"),
				Sprintf(`Red{*multi*}
Green{_line_}
Blue{~content~}
`)
		}

		It("should preserve already existing colors and text emphasis", func() {
			headline, content := setupTestStrings()
			Expect("\n" + ContentBox(headline, content)).To(BeEquivalentTo(Sprintf(`
╭ CornflowerBlue{~headline~}
│ Red{*multi*}
│ Green{_line_}
│ Blue{~content~}
╵
`)))
		})

		It("should overwrite existing headline color if it is specified", func() {
			headline, content := setupTestStrings()
			Expect("\n" + ContentBox(headline, content,
				HeadlineColor(DimGray),
			)).To(BeEquivalentTo(Sprintf(`
DimGray{╭} DimGray{~headline~}
DimGray{│} Red{*multi*}
DimGray{│} Green{_line_}
DimGray{│} Blue{~content~}
DimGray{╵}
`)))
		})

		It("should overwrite existing content color if it is specified", func() {
			headline, content := setupTestStrings()
			Expect("\n" + ContentBox(headline, content,
				ContentColor(DimGray),
			)).To(BeEquivalentTo(Sprintf(`
╭ CornflowerBlue{~headline~}
│ DimGray{*multi*}
│ DimGray{_line_}
│ DimGray{~content~}
╵
`)))
		})
	})

	Context("using reader directly", func() {
		It("should create a box using a reader", func() {
			r, w := io.Pipe()
			go func() {
				w.Write([]byte("multi\n"))
				w.Write([]byte("line\n"))
				w.Write([]byte("content\n"))
				w.Close()
			}()

			var buf bytes.Buffer
			Box(&buf, "headline", r)

			Expect("\n" + buf.String()).To(BeEquivalentTo(Sprintf(`
╭ headline
│ multi
│ line
│ content
╵
`)))
		})

		It("should create no box if no content could be obtained from the reader", func() {
			r, w := io.Pipe()
			go func() {
				w.Close()
			}()

			var buf bytes.Buffer
			Box(&buf, "headline", r)

			Expect(len(buf.String())).To(BeIdenticalTo(0))
		})
	})

	Context("using line wrap", func() {
		var tmp int

		BeforeEach(func() {
			tmp = FixedTerminalWidth
			FixedTerminalWidth = 80
		})

		AfterEach(func() {
			FixedTerminalWidth = tmp
		})

		It("should wrap lines that are too long", func() {
			Expect("\n" + ContentBox(
				"headline",
				"content with a very long first line, that is most likely an error message with a lot of context or similar",
			)).To(BeEquivalentTo(Sprintf(`
╭ headline
│ content with a very long first line, that is most likely an error message
│ with a lot of context or similar
╵
`)))
		})

		It("should not wrap long lines if wrapping is disabled", func() {
			Expect("\n" + ContentBox(
				"headline",
				"content with a very long first line, that is most likely an error message with a lot of context or similar",
				NoLineWrap(),
			)).To(BeEquivalentTo(Sprintf(`
╭ headline
│ content with a very long first line, that is most likely an error message with a lot of context or similar
╵
`)))
		})

		It("should not fail with empty lines", func() {
			Expect("\n" + ContentBox(
				"headline",
				" ",
			)).To(BeEquivalentTo(Sprintf(`
╭ headline
│  
╵
`)))
		})
	})
})
