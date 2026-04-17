package spruce

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/starkandwayne/goutils/tree"
)

var _ = Describe("Sort", func() {
	Describe("SortOperator.Run", func() {
		It("returns an error indicating orphaned operator", func() {
			op := &SortOperator{}
			ev := &Evaluator{Here: &tree.Cursor{Nodes: []string{"foobar"}}}

			resp, err := op.Run(ev, []*Expr{})
			Expect(resp).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("orphaned (( sort )) operator at $.foobar, no list exists at that path"))
		})
	})

	Describe("sortList", func() {
		It("returns an empty list unchanged", func() {
			list := []interface{}{}
			err := sortList("some.path", list, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(BeEmpty())
		})

		It("sorts integers", func() {
			list := []interface{}{2, 1}
			err := sortList("some.path", list, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(Equal([]interface{}{1, 2}))
		})

		It("sorts floats", func() {
			list := []interface{}{2.0, 1.0}
			err := sortList("some.path", list, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(Equal([]interface{}{1.0, 2.0}))
		})

		It("sorts strings", func() {
			list := []interface{}{"spruce", "spiff"}
			err := sortList("some.path", list, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(Equal([]interface{}{"spiff", "spruce"}))
		})

		It("sorts named-entry lists by name key", func() {
			list := []interface{}{
				map[interface{}]interface{}{"name": "B"},
				map[interface{}]interface{}{"name": "A"},
			}
			err := sortList("some.path", list, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(list).To(Equal([]interface{}{
				map[interface{}]interface{}{"name": "A"},
				map[interface{}]interface{}{"name": "B"},
			}))
		})

		It("fails on lists of lists", func() {
			list := []interface{}{
				[]interface{}{"B", "A"},
				[]interface{}{"A", "B"},
			}
			err := sortList("some.path", list, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("$.some.path is a list with list entries (not a list with maps, strings or numbers)"))
		})

		It("fails on inhomogeneous types", func() {
			list := []interface{}{42, 42.0, "42"}
			err := sortList("some.path", list, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("$.some.path is a list with different types (not a list with homogeneous entry types)"))
		})

		It("fails on nil values in list", func() {
			list := []interface{}{"A", "B", "C", nil}
			err := sortList("some.path", list, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("$.some.path is a list with different types (not a list with homogeneous entry types)"))
		})

		It("fails on inconsistent identifier key", func() {
			list := []interface{}{
				map[interface{}]interface{}{"foo": "one"},
				map[interface{}]interface{}{"key": "two"},
			}
			err := sortList("some.path", list, "foo")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("$.some.path is a list with map entries, where some do not contain foo (not a list with map entries each containing foo)"))
		})
	})
})
