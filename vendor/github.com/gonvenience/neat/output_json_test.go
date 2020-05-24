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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	yamlv2 "gopkg.in/yaml.v2"

	. "github.com/gonvenience/neat"
)

var _ = Describe("JSON output", func() {
	Context("create JSON output", func() {
		It("should create JSON output for a simple list", func() {
			result, err := ToJSONString([]interface{}{
				"one",
				"two",
				"three",
			})

			Expect(err).To(BeNil())
			Expect(result).To(BeEquivalentTo(`["one", "two", "three"]`))
		})

		It("should create JSON output of nested maps", func() {
			example := yamlv2.MapSlice{
				yamlv2.MapItem{
					Key: "map",
					Value: yamlv2.MapSlice{
						yamlv2.MapItem{
							Key: "foo",
							Value: yamlv2.MapSlice{
								yamlv2.MapItem{
									Key: "bar",
									Value: yamlv2.MapSlice{
										yamlv2.MapItem{
											Key:   "name",
											Value: "foobar",
										},
									},
								},
							},
						},
					},
				},
			}

			var result string
			var err error
			var outputProcessor = NewOutputProcessor(false, false, &DefaultColorSchema)

			result, err = outputProcessor.ToCompactJSON(example)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeEquivalentTo(`{"map": {"foo": {"bar": {"name": "foobar"}}}}`))

			result, err = outputProcessor.ToJSON(example)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeEquivalentTo(`{
  "map": {
    "foo": {
      "bar": {
        "name": "foobar"
      }
    }
  }
}`))
		})

		It("should create JSON output for empty structures", func() {
			example := yamlv2.MapSlice{
				yamlv2.MapItem{
					Key:   "empty-map",
					Value: yamlv2.MapSlice{},
				},

				yamlv2.MapItem{
					Key:   "empty-list",
					Value: []interface{}{},
				},

				yamlv2.MapItem{
					Key:   "empty-scalar",
					Value: nil,
				},
			}

			var result string
			var err error
			var outputProcessor = NewOutputProcessor(false, false, &DefaultColorSchema)

			result, err = outputProcessor.ToCompactJSON(example)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeEquivalentTo(`{"empty-map": {}, "empty-list": [], "empty-scalar": null}`))

			result, err = outputProcessor.ToJSON(example)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(BeEquivalentTo(`{
  "empty-map": {},
  "empty-list": [],
  "empty-scalar": null
}`))
		})
	})
})
