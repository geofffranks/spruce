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

package ytbx_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gonvenience/ytbx"
)

var _ = Describe("Restructure order of map keys", func() {
	Context("YAML MapSlice key reorderings of the MapSlice itself", func() {
		It("should restructure Concourse root level keys", func() {
			example := yml("{ groups: [], jobs: [], resources: [], resource_types: [] }")
			RestructureObject(example)

			keys, err := ListStringKeys(example)
			Expect(err).ToNot(HaveOccurred())
			Expect(keys).To(BeEquivalentTo([]string{"jobs", "resources", "resource_types", "groups"}))
		})

		It("should restructure Concourse resource and resource_type keys", func() {
			example := yml("{ source: {}, name: {}, type: {}, privileged: {} }")
			RestructureObject(example)

			keys, err := ListStringKeys(example)
			Expect(err).ToNot(HaveOccurred())
			Expect(keys).To(BeEquivalentTo([]string{"name", "type", "source", "privileged"}))
		})
	})

	Context("YAML MapSlice key reorderings of the MapSlice values", func() {
		It("should restructure Concourse resource keys as part as part of a MapSlice value", func() {
			example := yml("{ resources: [ { privileged: false, source: { branch: foo, paths: [] }, name: myname, type: mytype } ] }")
			RestructureObject(example)

			keys, err := ListStringKeys(example.Content[1].Content[0])
			Expect(err).ToNot(HaveOccurred())
			Expect(keys).To(BeEquivalentTo([]string{"name", "type", "source", "privileged"}))
		})
	})

	Context("Restructure code tries to rearrange even unknown keys", func() {
		It("should reorder map keys in a somehow more readable way", func() {
			example := yml(`{"list":["one","two","three"], "some":{"deep":{"structure":{"where":{"you":{"loose":{"focus":{"one":1,"two":2}}}}}}}, "name":"here", "release":"this"}`)
			RestructureObject(example)

			keys, err := ListStringKeys(example)
			Expect(err).ToNot(HaveOccurred())
			Expect(keys).To(BeEquivalentTo([]string{"name", "release", "list", "some"}))
		})
	})
})
