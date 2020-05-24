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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gonvenience/neat"
)

var _ = Describe("Table formatting", func() {
	Context("Process two-dimensional slices as tables to be printed", func() {
		It("should work for simplest of tables with no additional formatting", func() {
			input := [][]string{
				{"eins", "zwei", "drei"},
				{"one", "two", "three"},
				{"un", "deux", "trois"},
			}

			expectedResult := `eins zwei drei
one  two  three
un   deux trois
`

			tableString, err := Table(input)
			Expect(err).ToNot(HaveOccurred())
			Expect(tableString).To(BeEquivalentTo(expectedResult))
		})

		It("should work with additional formatting for inner borders", func() {
			input := [][]string{
				{"eins", "zwei", "drei"},
				{"one", "two", "three"},
				{"un", "deux", "trois"},
			}

			expectedResult := `eins │ zwei │ drei
one  │ two  │ three
un   │ deux │ trois
`

			tableString, err := Table(input, VertialBarSeparator())
			Expect(err).ToNot(HaveOccurred())
			Expect(tableString).To(BeEquivalentTo(expectedResult))
		})

		It("should work with a custom separator string", func() {
			input := [][]string{
				{"eins", "zwei", "drei"},
				{"one", "two", "three"},
				{"un", "deux", "trois"},
			}

			expectedResult := `eins / zwei / drei
one  / two  / three
un   / deux / trois
`

			tableString, err := Table(input, CustomSeparator(" / "))
			Expect(err).ToNot(HaveOccurred())
			Expect(tableString).To(BeEquivalentTo(expectedResult))
		})

		It("should work with additional formatting for row padding", func() {
			input := [][]string{
				{"eins", "zwei", "drei"},
				{"one", "two", "three"},
				{"un", "deux", "trois"},
			}

			expectedResult := `eins                      zwei                      drei
one                       two                       three
un                        deux                      trois
`

			tableString, err := Table(input, DesiredWidth(80))
			Expect(err).ToNot(HaveOccurred())
			Expect(tableString).To(BeEquivalentTo(expectedResult))
		})

		It("should work with additional formatting for alignment", func() {
			input := [][]string{
				{"eins", "zwei", "drei", "vier", "fünf"},
				{"one", "two", "three", "four", "five"},
				{"un", "deux", "trois", "quatre", "cinq"},
				{"uno", "dos", "tres", "cuatro", "cinco"},
			}

			expectedResult := `eins zwei drei   vier   fünf
one   two three  four   five
un   deux trois quatre  cinq
uno   dos tres  cuatro cinco
`

			tableString, err := Table(input, AlignRight(1, 4), AlignCenter(3))
			Expect(err).ToNot(HaveOccurred())
			Expect(fmt.Sprintf("%#v", tableString)).To(BeEquivalentTo(fmt.Sprintf("%#v", expectedResult)))
		})

		It("should error if a row would exceed the desired table width", func() {
			input := [][]string{
				{"#1", "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum."},
				{"#2", "Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet."},
			}

			tableString, err := Table(input, DesiredWidth(120))
			Expect(err).Should(MatchError(&RowLengthExceedsDesiredWidthError{}))
			Expect(tableString).To(BeEquivalentTo(""))
		})

		It("should error if empty input is provided", func() {
			tableString, err := Table(nil)
			Expect(err).Should(MatchError(&EmptyTableError{}))
			Expect(tableString).To(BeEquivalentTo(""))
		})

		It("should error if imbalanced table is provided", func() {
			tableString, err := Table([][]string{
				{"eins", "zwei", "drei", "vier", "fünf"},
				{"one", "two", "three", "four"},
			})
			Expect(err).Should(MatchError(&ImbalancedTableError{}))
			Expect(tableString).To(BeEquivalentTo(""))
		})

		It("should error if a column index based table option is out of bounds", func() {
			tableString, err := Table([][]string{
				{"eins", "zwei", "drei", "vier"},
				{"one", "two", "three", "four"},
			}, AlignCenter(4))
			Expect(err).Should(MatchError(&ColumnIndexIsOutOfBoundsError{4}))
			Expect(tableString).To(BeEquivalentTo(""))
		})

		It("should be possible to create a table without a final linefeed", func() {
			input := [][]string{
				{"eins", "zwei", "drei"},
				{"one", "two", "three"},
				{"un", "deux", "trois"},
			}

			expectedResult := `eins zwei drei
one  two  three
un   deux trois`

			tableString, err := Table(input, OmitLinefeedAtTableEnd())
			Expect(err).ToNot(HaveOccurred())
			Expect(tableString).To(BeEquivalentTo(expectedResult))
		})

		It("should be possible to limit the number of rows", func() {
			input := [][]string{
				{"eins", "zwei", "drei"},
				{"one", "two", "three"},
				{"un", "deux", "trois"},
			}

			expectedResult := `eins zwei drei
one  two  three
[...]
`

			tableString, err := Table(input, LimitRows(2))
			Expect(err).ToNot(HaveOccurred())
			Expect(tableString).To(BeEquivalentTo(expectedResult))
		})

		It("should not fail when using a row limit which is greater than the table length", func() {
			input := [][]string{
				{"eins", "zwei", "drei"},
				{"one", "two", "three"},
				{"un", "deux", "trois"},
			}

			expectedResult := `eins zwei drei
one  two  three
un   deux trois
`

			tableString, err := Table(input, LimitRows(25))
			Expect(err).ToNot(HaveOccurred())
			Expect(tableString).To(BeEquivalentTo(expectedResult))
		})
	})
})
