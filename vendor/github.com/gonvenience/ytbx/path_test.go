// Copyright © 2018 The Homeport Team
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
	yamlv3 "gopkg.in/yaml.v3"
)

func getExampleDocument() *yamlv3.Node {
	input, err := LoadFile(assets("testbed", "example.yml"))
	Expect(err).ToNot(HaveOccurred())
	Expect(len(input.Documents)).To(BeIdenticalTo(1))

	return input.Documents[0]
}

var _ = Describe("path tests", func() {
	Context("parse dot-style path strings into a path", func() {
		It("should parse string with only map elements", func() {
			path, err := ParseDotStylePathString("yaml.structure.somekey", getExampleDocument())
			Expect(err).To(BeNil())
			Expect(path).To(BeEquivalentTo(Path{DocumentIdx: 0, PathElements: []PathElement{
				{Idx: -1, Key: "", Name: "yaml"},
				{Idx: -1, Key: "", Name: "structure"},
				{Idx: -1, Key: "", Name: "somekey"},
			}}))
		})

		It("should parse string with map and named-entry list elements", func() {
			path, err := ParseDotStylePathString("list.one.somekey", getExampleDocument())
			Expect(err).To(BeNil())
			Expect(path).To(BeEquivalentTo(Path{DocumentIdx: 0, PathElements: []PathElement{
				{Idx: -1, Key: "", Name: "list"},
				{Idx: -1, Key: "name", Name: "one"},
				{Idx: -1, Key: "", Name: "somekey"},
			}}))
		})

		It("should parse string with simple list entry", func() {
			path, err := ParseDotStylePathString("simpleList.1", getExampleDocument())
			Expect(err).To(BeNil())
			Expect(path).To(BeEquivalentTo(Path{DocumentIdx: 0, PathElements: []PathElement{
				{Idx: -1, Key: "", Name: "simpleList"},
				{Idx: 1},
			}}))
		})

		It("should parse string with non-existing map elements", func() {
			path, err := ParseDotStylePathString("yaml.update.newkey", getExampleDocument())
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(BeEquivalentTo(Path{DocumentIdx: 0, PathElements: []PathElement{
				{Idx: -1, Key: "", Name: "yaml"},
				{Idx: -1, Key: "", Name: "update"},
				{Idx: -1, Key: "", Name: "newkey"},
			}}))
		})

		It("should parse string with non-existing map and named-entry list elements", func() {
			path, err := ParseDotStylePathString("list.one.newkey", getExampleDocument())
			Expect(err).To(BeNil())
			Expect(path).To(BeEquivalentTo(Path{DocumentIdx: 0, PathElements: []PathElement{
				{Idx: -1, Key: "", Name: "list"},
				{Idx: -1, Key: "name", Name: "one"},
				{Idx: -1, Key: "", Name: "newkey"},
			}}))
		})
	})

	Context("parse go-patch style path strings into paths", func() {
		It("should parse an input string using go-patch style into a path (only maps)", func() {
			path, err := ParseGoPatchStylePathString("/yaml/structure/somekey")
			Expect(err).To(BeNil())
			Expect(path).To(BeEquivalentTo(Path{DocumentIdx: 0, PathElements: []PathElement{
				{Idx: -1, Key: "", Name: "yaml"},
				{Idx: -1, Key: "", Name: "structure"},
				{Idx: -1, Key: "", Name: "somekey"},
			}}))
		})

		It("should parse an input string using go-patch style into a path (maps and named-entry lists)", func() {
			path, err := ParseGoPatchStylePathString("/list/name=one/somekey")
			Expect(err).To(BeNil())
			Expect(path).To(BeEquivalentTo(Path{DocumentIdx: 0, PathElements: []PathElement{
				{Idx: -1, Key: "", Name: "list"},
				{Idx: -1, Key: "name", Name: "one"},
				{Idx: -1, Key: "", Name: "somekey"},
			}}))
		})

		It("should parse an input string using go-patch style into a path (simple list)", func() {
			path, err := ParseGoPatchStylePathString("/simpleList/1")
			Expect(err).To(BeNil())
			Expect(path).To(BeEquivalentTo(Path{DocumentIdx: 0, PathElements: []PathElement{
				{Idx: -1, Key: "", Name: "simpleList"},
				{Idx: 1},
			}}))
		})

		It("should parse an input string that points to the root of the tree structure", func() {
			path, err := ParseGoPatchStylePathString("/")
			Expect(err).To(BeNil())
			Expect(path).To(BeEquivalentTo(Path{DocumentIdx: 0, PathElements: nil}))
		})

		It("should parse real-life scenario paths with mixed types", func() {
			path, err := ParseGoPatchStylePathString("/resource_pools/name=concourse_resource_pool/cloud_properties/datacenters/0/clusters")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(BeEquivalentTo(Path{DocumentIdx: 0, PathElements: []PathElement{
				{Idx: -1, Key: "", Name: "resource_pools"},
				{Idx: -1, Key: "name", Name: "concourse_resource_pool"},
				{Idx: -1, Key: "", Name: "cloud_properties"},
				{Idx: -1, Key: "", Name: "datacenters"},
				{Idx: 0},
				{Idx: -1, Key: "", Name: "clusters"},
			}}))
		})

		It("should parse path strings with escaped slashes", func() {
			path, err := ParseGoPatchStylePathString("/foo/name=bar.com\\/id/string")
			Expect(err).ToNot(HaveOccurred())
			Expect(path).To(BeEquivalentTo(Path{DocumentIdx: 0, PathElements: []PathElement{
				{Idx: -1, Key: "", Name: "foo"},
				{Idx: -1, Key: "name", Name: "bar.com/id"},
				{Idx: -1, Key: "", Name: "string"},
			}}))
		})
	})

	Context("compare paths between two files", func() {
		It("should find only duplicate paths", func() {
			list, err := ComparePaths(assets("testbed", "sample_a.yml"), assets("testbed", "sample_b.yml"), GoPatchStyle, false)
			Expect(err).ToNot(HaveOccurred())

			listOfPaths := []Path{
				{
					DocumentIdx: 0, PathElements: []PathElement{
						{Idx: -1, Key: "", Name: "yaml"},
						{Idx: -1, Key: "", Name: "structure"},
						{Idx: -1, Key: "", Name: "somekey"},
					},
				},
				{
					DocumentIdx: 0, PathElements: []PathElement{
						{Idx: -1, Key: "", Name: "yaml"},
						{Idx: -1, Key: "", Name: "structure"},
						{Idx: -1, Key: "", Name: "dot"},
					},
				},
				{
					DocumentIdx: 0, PathElements: []PathElement{
						{Idx: -1, Key: "", Name: "list"},
						{Idx: -1, Key: "name", Name: "sametwo"},
						{Idx: -1, Key: "", Name: "somekey"},
					},
				},
			}

			Expect(list).To(BeEquivalentTo(listOfPaths))
		})

		It("should find only paths with the same value", func() {
			list, err := ComparePaths(assets("testbed", "sample_a.yml"), assets("testbed", "sample_b.yml"), GoPatchStyle, true)
			Expect(err).ToNot(HaveOccurred())

			listOfPathsWithSameValue := []Path{
				{
					DocumentIdx: 0, PathElements: []PathElement{
						{Idx: -1, Key: "", Name: "yaml"},
						{Idx: -1, Key: "", Name: "structure"},
						{Idx: -1, Key: "", Name: "dot"},
					},
				},
				{
					DocumentIdx: 0, PathElements: []PathElement{
						{Idx: -1, Key: "", Name: "list"},
						{Idx: -1, Key: "name", Name: "sametwo"},
						{Idx: -1, Key: "", Name: "somekey"},
					},
				},
			}

			Expect(list).To(BeEquivalentTo(listOfPathsWithSameValue))
		})
	})
})
