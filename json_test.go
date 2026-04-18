package spruce

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSON", func() {
	Describe("jsonifyData", func() {
		It("converts valid YAML to JSON", func() {
			data := []byte("key: value\n")
			result, err := jsonifyData(data, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(`{"key":"value"}`))
		})

		It("returns an error for invalid YAML", func() {
			data := []byte(":\n  - :\n    -")
			_, err := jsonifyData(data, false)
			Expect(err).To(HaveOccurred())
		})

		It("returns an error when root is not a map", func() {
			data := []byte("- one\n- two\n")
			_, err := jsonifyData(data, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Root of YAML document is not a hash/map"))
		})

		It("converts nested structures", func() {
			data := []byte("top:\n  sub: val\n  list:\n  - a\n  - b\n")
			result, err := jsonifyData(data, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring(`"top"`))
			Expect(result).To(ContainSubstring(`"sub":"val"`))
		})
	})

	Describe("JSONifyIO", func() {
		It("converts YAML from a reader to JSON", func() {
			reader := strings.NewReader("key: value\n")
			result, err := JSONifyIO(reader, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(`{"key":"value"}`))
		})

		It("returns an error for invalid YAML from reader", func() {
			reader := strings.NewReader("- not a map\n")
			_, err := JSONifyIO(reader, false)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("deinterface", func() {
		It("passes through scalars unchanged", func() {
			result, err := deinterface("hello", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("hello"))
		})

		It("converts map[interface{}]interface{} to map[string]interface{}", func() {
			input := map[interface{}]interface{}{"key": "value"}
			result, err := deinterface(input, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(map[string]interface{}{"key": "value"}))
		})

		It("converts nested lists", func() {
			input := []interface{}{"a", "b"}
			result, err := deinterface(input, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal([]interface{}{"a", "b"}))
		})
	})

	Describe("deinterfaceMap", func() {
		Context("in strict mode", func() {
			It("returns an error for non-string keys", func() {
				input := map[interface{}]interface{}{42: "value"}
				_, err := deinterfaceMap(input, true)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("non-string keys"))
			})
		})

		Context("in non-strict mode", func() {
			It("coerces non-string keys to strings", func() {
				input := map[interface{}]interface{}{42: "value"}
				result, err := deinterfaceMap(input, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveKey("42"))
			})
		})
	})

	Describe("addKeyToMap", func() {
		It("adds a key-value pair", func() {
			m := map[string]interface{}{}
			err := addKeyToMap(m, "key", "value", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(m).To(HaveKeyWithValue("key", "value"))
		})

		It("warns on duplicate keys", func() {
			m := map[string]interface{}{"key": "old"}
			err := addKeyToMap(m, "key", "new", false)
			Expect(err).NotTo(HaveOccurred())
			// Duplicate detected — original value preserved
			Expect(m["key"]).To(Equal("old"))
		})
	})

	Describe("JSONifyFiles", func() {
		var tmpFile *os.File

		BeforeEach(func() {
			var err error
			tmpFile, err = os.CreateTemp("", "spruce-json-test-*.yml")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			os.Remove(tmpFile.Name())
		})

		It("converts a YAML file to JSON", func() {
			_, err := tmpFile.WriteString("key: value\n")
			Expect(err).NotTo(HaveOccurred())
			Expect(tmpFile.Close()).To(Succeed())

			results, err := JSONifyFiles([]string{tmpFile.Name()}, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(1))
			Expect(results[0]).To(Equal(`{"key":"value"}`))
		})

		It("returns an error for a non-existent file", func() {
			_, err := JSONifyFiles([]string{"/nonexistent/path/file.yml"}, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Error reading file"))
		})

		It("handles multiple YAML documents in a single file", func() {
			_, err := tmpFile.WriteString("a: 1\n---\nb: 2\n")
			Expect(err).NotTo(HaveOccurred())
			Expect(tmpFile.Close()).To(Succeed())

			results, err := JSONifyFiles([]string{tmpFile.Name()}, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(2))
			Expect(results[0]).To(Equal(`{"a":1}`))
			Expect(results[1]).To(Equal(`{"b":2}`))
		})

		It("converts multiple files", func() {
			_, err := tmpFile.WriteString("first: 10\n")
			Expect(err).NotTo(HaveOccurred())
			Expect(tmpFile.Close()).To(Succeed())

			tmpFile2, err2 := os.CreateTemp("", "spruce-json-test2-*.yml")
			Expect(err2).NotTo(HaveOccurred())
			defer os.Remove(tmpFile2.Name())
			_, err = tmpFile2.WriteString("second: 20\n")
			Expect(err).NotTo(HaveOccurred())
			Expect(tmpFile2.Close()).To(Succeed())

			results, err := JSONifyFiles([]string{tmpFile.Name(), tmpFile2.Name()}, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(results).To(HaveLen(2))
			Expect(results[0]).To(Equal(`{"first":10}`))
			Expect(results[1]).To(Equal(`{"second":20}`))
		})
	})
})
