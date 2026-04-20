package spruce

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("getIndexOfSimpleEntry", func() {
	It("finds the index of a simple value in a list", func() {
		list := []interface{}{"a", "b", "c"}
		idx := getIndexOfSimpleEntry(list, "b")
		Expect(idx).To(Equal(1))
	})

	It("returns -1 when the value is not found", func() {
		list := []interface{}{"a", "b", "c"}
		idx := getIndexOfSimpleEntry(list, "z")
		Expect(idx).To(Equal(-1))
	})

	It("returns -1 for an empty list", func() {
		list := []interface{}{}
		idx := getIndexOfSimpleEntry(list, "a")
		Expect(idx).To(Equal(-1))
	})

	It("skips non-string entries", func() {
		list := []interface{}{42, "target", true}
		idx := getIndexOfSimpleEntry(list, "target")
		Expect(idx).To(Equal(1))
	})
})
