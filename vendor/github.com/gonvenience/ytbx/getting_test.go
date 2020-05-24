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

	"github.com/gonvenience/ytbx"
)

var _ = Describe("getting stuff test cases", func() {
	Context("Grabing values by path", func() {
		It("should return the value referenced by the path", func() {
			example := yml(assets("examples", "types.yml"))
			Expect(grab(example, "/yaml/map/before")).To(BeEquivalentTo("after"))
			Expect(grab(example, "/yaml/map/intA")).To(BeEquivalentTo(42))
			Expect(grab(example, "/yaml/map/mapA")).To(BeAsNode(yml(`{ key0: A, key1: A }`)))
			Expect(grab(example, "/yaml/map/listA")).To(BeAsNode(list(`[ A, A, A ]`)))
			Expect(grab(example, "/yaml/named-entry-list-using-name/name=B")).To(BeAsNode(yml(`{ name: B }`)))
			Expect(grab(example, "/yaml/named-entry-list-using-key/key=B")).To(BeAsNode(yml(`{ key: B }`)))
			Expect(grab(example, "/yaml/named-entry-list-using-id/id=B")).To(BeAsNode(yml(`{ id: B }`)))
			Expect(grab(example, "/yaml/simple-list/1")).To(BeEquivalentTo("B"))
			Expect(grab(example, "/yaml/named-entry-list-using-key/3")).To(BeAsNode(yml(`{ key: X }`)))

			example = yml(assets("bosh-yaml", "manifest.yml"))
			Expect(grab(example, "/instance_groups/name=web/networks/name=concourse/static_ips/0")).To(BeEquivalentTo("XX.XX.XX.XX"))
			Expect(grab(example, "/instance_groups/name=worker/jobs/name=baggageclaim/properties")).To(BeAsNode(yml(`{}`)))
		})

		It("should return the whole tree if root is referenced", func() {
			file, err := ytbx.LoadFile(assets("examples", "types.yml"))
			Expect(err).ToNot(HaveOccurred())

			document := file.Documents[0]
			Expect(grab(document, "/")).To(BeAsNode(document.Content[0]))
		})

		It("should return useful error messages", func() {
			example := yml(assets("examples", "types.yml"))
			Expect(grabError(example, "/yaml/simple-list/-1")).To(BeEquivalentTo("failed to traverse tree, provided list index -1 is not in range: 0..4"))
			Expect(grabError(example, "/yaml/does-not-exist")).To(BeEquivalentTo("no key 'does-not-exist' found in map, available keys: map, simple-list, named-entry-list-using-name, named-entry-list-using-key, named-entry-list-using-id"))
			Expect(grabError(example, "/yaml/0")).To(BeEquivalentTo("failed to traverse tree, expected list but found type map at /yaml"))
			Expect(grabError(example, "/yaml/simple-list/foobar")).To(BeEquivalentTo("failed to traverse tree, expected map but found type list at /yaml/simple-list"))
			Expect(grabError(example, "/yaml/map/foobar=0")).To(BeEquivalentTo("failed to traverse tree, expected complex-list but found type map at /yaml/map"))
			Expect(grabError(example, "/yaml/named-entry-list-using-id/id=0")).To(BeEquivalentTo("there is no entry id=0 in the list"))
		})
	})

	Context("Trying to get values by path in an empty file", func() {
		It("should return a not found key error", func() {
			emptyFile := yml(assets("examples", "empty.yml"))
			Expect(grabError(emptyFile, "/does-not-exist")).To(
				BeEquivalentTo("failed to traverse tree, expected map but found type string at /"),
			)
		})
	})
})
