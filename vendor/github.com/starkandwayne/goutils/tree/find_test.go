package tree_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/starkandwayne/goutils/tree"
)

var _ = Describe("Tree Walking", func() {

	Context("tree.Find() Functions", func() {
		data := map[string]interface{}{
			"string":  "asdf",
			"number":  1234,
			"boolean": true,
			"map": map[string]interface{}{
				"k": "v",
				"n": 1,
			},
			"array": []interface{}{
				1,
				2,
				"fdsa",
			},
		}

		Context("tree.FindString()", func() {
			It("should fail if tree.Find() fails", func() {
				v, err := tree.FindString(data, "nonexistent")
				Ω(err).Should(HaveOccurred())
				Ω(v).Should(Equal(""))
			})

			It("should fail if path is not a string", func() {
				v, err := tree.FindString(data, "number")
				Ω(err).Should(HaveOccurred())
				Ω(v).Should(Equal(""))
			})

			It("should succeed if path exists and is string", func() {
				v, err := tree.FindString(data, "string")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal("asdf"))
			})
		})

		Context("tree.FindNum()", func() {
			It("should fail if tree.Find() fails", func() {
				v, err := tree.FindNum(data, "nonexistent")
				Ω(err).Should(HaveOccurred())
				Ω(v).Should(Equal(tree.Number(0)))
			})

			It("should fail if path is not a number", func() {
				v, err := tree.FindNum(data, "string")
				Ω(err).Should(HaveOccurred())
				Ω(v).Should(Equal(tree.Number(0)))
			})

			It("should succeed if path exists and is numeric", func() {
				v, err := tree.FindNum(data, "number")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal(tree.Number(1234)))
			})
		})

		Context("tree.FindBool()", func() {
			It("should fail if tree.Find() fails", func() {
				v, err := tree.FindBool(data, "nonexistent")
				Ω(err).Should(HaveOccurred())
				Ω(v).Should(Equal(false))
			})

			It("should fail if path is not a bool", func() {
				v, err := tree.FindBool(data, "number")
				Ω(err).Should(HaveOccurred())
				Ω(v).Should(Equal(false))
			})

			It("should succeed if path exists and is boolean", func() {
				v, err := tree.FindBool(data, "boolean")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal(true))
			})
		})

		Context("tree.FindMap()", func() {
			It("should fail if tree.Find() fails", func() {
				v, err := tree.FindMap(data, "nonexistent")
				Ω(err).Should(HaveOccurred())
				Ω(v).Should(Equal(map[string]interface{}{}))
			})

			It("should fail if path is not a map", func() {
				v, err := tree.FindMap(data, "number")
				Ω(err).Should(HaveOccurred())
				Ω(v).Should(Equal(map[string]interface{}{}))
			})

			It("should succeed if path exists and is a map", func() {
				v, err := tree.FindMap(data, "map")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal(map[string]interface{}{
					"k": "v",
					"n": 1,
				}))

				n, err := tree.FindString(data, "map.k")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(n).Should(Equal("v"))
			})
		})

		Context("tree.FindArray()", func() {
			It("should fail if tree.Find() fails", func() {
				v, err := tree.FindArray(data, "nonexistent")
				Ω(err).Should(HaveOccurred())
				Ω(v).Should(Equal([]interface{}{}))
			})

			It("should fail if path is not an array", func() {
				v, err := tree.FindArray(data, "number")
				Ω(err).Should(HaveOccurred())
				Ω(v).Should(Equal([]interface{}{}))
			})

			It("should succeed if path exists and is an array", func() {
				v, err := tree.FindArray(data, "array")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal([]interface{}{
					1,
					2,
					"fdsa",
				}))

				f, err := tree.FindString(data, "array.[2]")
				Ω(f).Should(Equal("fdsa"))
			})
		})

		Context("tree.Find()", func() {
			It("should fail if tree.Find() fails", func() {
				v, err := tree.Find(data, "nonexistent")
				Ω(err).Should(HaveOccurred())
				Ω(v).Should(BeNil())
			})

			It("should succeed if path exists", func() {
				v, err := tree.Find(data, "string")
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal(interface{}("asdf")))
			})
		})
	})

	Context("tree.Numbers", func() {
		Context("Int64()", func() {
			It("fails if the value is not an integer", func() {
				n := tree.Number(1.1)
				i, err := n.Int64()
				Ω(i).Should(Equal(int64(0)))
				Ω(err).Should(HaveOccurred())
			})

			It("succeeds if the value is an integer", func() {
				n := tree.Number(1)
				i, err := n.Int64()
				Ω(i).Should(Equal(int64(1)))
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

		Context("Float64()", func() {
			It("returns a float64 cast of the value", func() {
				n := tree.Number(1.1)
				f := n.Float64()
				Ω(f).Should(Equal(1.1))
			})
		})

		Context("String()", func() {
			It("returns a float-string representation of non-integers", func() {
				n := tree.Number(1.1)
				s := n.String()
				Ω(s).Should(Equal("1.100000"))
			})

			It("returns an int-string representation of integers", func() {
				n := tree.Number(1)
				s := n.String()
				Ω(s).Should(Equal("1"))
			})
		})
	})
})
