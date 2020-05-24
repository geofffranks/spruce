package botta_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/starkandwayne/goutils/botta"
	"github.com/starkandwayne/goutils/tree"
)

var _ = Describe("Response Obj", func() {
	httpResponse := &http.Response{}
	response := botta.Response{
		HTTPResponse: httpResponse,
		Raw:          []byte(`{"json":"content returned from server"}`),
		Data: map[string]interface{}{
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
		},
	}
	Context("Response", func() {
		It("should have publicly accessible Raw response", func() {
			Expect(response.Raw).Should(Equal([]byte(`{"json":"content returned from server"}`)))
		})
		It("should have publicly accessible pointer to the HTTP response", func() {
			Expect(response.HTTPResponse).Should(Equal(httpResponse))
		})
	})

	Context("Response.StringVal()", func() {
		It("should fail when specified path does not point to a string", func() {
			str, err := response.StringVal("number")
			Expect(err).Should(HaveOccurred())
			Expect(str).Should(Equal(""))
		})
		It("should succeed when specified path points to a string", func() {
			str, err := response.StringVal("string")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(str).Should(Equal("asdf"))
		})
	})

	Context("Response.NumVal()", func() {
		It("should fail when specified path does not point to a number", func() {
			num, err := response.NumVal("string")
			Expect(err).Should(HaveOccurred())
			Expect(num).Should(Equal(tree.Number(0)))
		})
		It("should succeed when specified path points to a number", func() {
			num, err := response.NumVal("number")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(num).Should(Equal(tree.Number(1234)))
		})
	})

	Context("Response.BoolVal()", func() {
		It("should fail when specified path does not point to a boolean", func() {
			b, err := response.BoolVal("number")
			Expect(err).Should(HaveOccurred())
			Expect(b).Should(Equal(false))
		})
		It("should succeed when specified path points to a boolean", func() {
			b, err := response.BoolVal("boolean")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(b).Should(BeTrue())
		})
	})

	Context("Response.MapVal()", func() {
		It("should fail when specified path does not point to a map", func() {
			m, err := response.MapVal("number")
			Expect(err).Should(HaveOccurred())
			Expect(m).Should(Equal(map[string]interface{}{}))
		})
		It("should succeed when specified path points to a m", func() {
			m, err := response.MapVal("map")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(m).Should(Equal(map[string]interface{}{
				"k": "v",
				"n": 1,
			}))

			n, err := response.StringVal("map.k")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(n).Should(Equal("v"))
		})
	})

	Context("Response.ArrayVal()", func() {
		It("should fail when specified path does not point to an array", func() {
			a, err := response.ArrayVal("number")
			Expect(err).Should(HaveOccurred())
			Expect(a).Should(Equal([]interface{}{}))
		})
		It("should succeed when specified path points to an array", func() {
			a, err := response.ArrayVal("array")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(a).Should(Equal([]interface{}{
				1,
				2,
				"fdsa",
			}))

			f, err := response.StringVal("array.[2]")
			Expect(f).Should(Equal("fdsa"))
		})
	})

	Context("Response.Val()", func() {
		It("should fail when specified path does not point to something", func() {
			i, err := response.Val("n'exist pas")
			Expect(err).Should(HaveOccurred())
			Expect(i).Should(BeNil())
		})
		It("should succeed when specified path points to something ", func() {
			i, err := response.Val("string")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(i).Should(Equal(interface{}("asdf")))
		})
	})
})
