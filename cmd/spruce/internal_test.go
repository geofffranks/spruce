package main

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// openFiles is a test helper used to open files for mergeAllDocs tests
func openFiles(paths []string) ([]YamlFile, error) {
	files := []YamlFile{}
	for _, file := range paths {
		f, err := os.Open(file)
		if err != nil {
			return files, err
		}
		files = append(files, YamlFile{Path: file, Reader: f})
	}
	return files, nil
}

var _ = Describe("parseYAML()", func() {
	It("returns error for invalid yaml data", func() {
		data := `
asdf: fdsa
- asdf: fdsa
`
		obj, err := parseYAML([]byte(data))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unmarshal []byte to yaml failed:"))
		Expect(obj).To(BeNil())
	})

	It("does not return error if yaml is empty", func() {
		data := `---
`
		obj, err := parseYAML([]byte(data))
		Expect(err).NotTo(HaveOccurred())
		Expect(obj).NotTo(BeNil())
	})

	It("returns error if yaml is a bool", func() {
		data := `
true
`
		obj, err := parseYAML([]byte(data))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Root of YAML document is not a hash/map:"))
		Expect(obj).To(BeNil())
	})

	It("returns error if yaml is a string", func() {
		data := `
"1234"
`
		obj, err := parseYAML([]byte(data))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Root of YAML document is not a hash/map:"))
		Expect(obj).To(BeNil())
	})

	It("returns error if yaml is a number", func() {
		data := `
1234
`
		obj, err := parseYAML([]byte(data))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Root of YAML document is not a hash/map:"))
		Expect(obj).To(BeNil())
	})

	It("returns error if yaml an array", func() {
		data := `
- 1
- 2
`
		obj, err := parseYAML([]byte(data))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Root of YAML document is not a hash/map:"))
		Expect(obj).To(BeNil())
	})

	It("returns expected datastructure from valid yaml", func() {
		data := `
top:
  subarray:
  - one
  - two
`
		obj, err := parseYAML([]byte(data))
		expect := map[interface{}]interface{}{
			"top": map[interface{}]interface{}{
				"subarray": []interface{}{"one", "two"},
			},
		}
		Expect(obj).To(Equal(expect))
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("mergeAllDocs()", func() {
	It("Fails with readFile error on bad first doc", func() {
		files, err := openFiles([]string{"../../assets/merge/second.yml"})
		files[0].Reader.Close()
		Expect(err).NotTo(HaveOccurred())
		_, err = mergeAllDocs(files, mergeOpts{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Error reading file ../../assets/merge/second.yml:"))
	})

	It("Fails with parseYAML error on bad second doc", func() {
		files, err := openFiles([]string{"../../assets/merge/first.yml", "../../assets/merge/bad.yml"})
		Expect(err).NotTo(HaveOccurred())
		_, err = mergeAllDocs(files, mergeOpts{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("../../assets/merge/bad.yml: Root of YAML document is not a hash/map:"))
	})

	It("Fails with mergeMap error", func() {
		files, err := openFiles([]string{"../../assets/merge/first.yml", "../../assets/merge/error.yml"})
		Expect(err).NotTo(HaveOccurred())
		_, err = mergeAllDocs(files, mergeOpts{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("$.array_inline.0: new object is a string, not a map - cannot merge by key"))
	})

	It("Succeeds with valid files + yaml", func() {
		expect := map[interface{}]interface{}{
			"key":           "overridden",
			"array_append":  []interface{}{"one", "two", "three"},
			"array_prepend": []interface{}{"three", "four", "five"},
			"array_replace": []interface{}{[]interface{}{1, 2, 3}},
			"array_inline": []interface{}{
				map[interface{}]interface{}{"name": "first_elem", "val": "overwritten"},
				"second_elem was overwritten",
				"third elem is appended",
			},
			"array_default": []interface{}{
				"FIRST",
				"SECOND",
				"third",
			},
			"array_map_default": []interface{}{
				map[interface{}]interface{}{
					"name": "AAA",
					"k1":   "key 1",
					"k2":   "updated",
				},
				map[interface{}]interface{}{
					"name": "BBB",
					"k2":   "final",
					"k3":   "original",
				},
			},
			"map": map[interface{}]interface{}{
				"key":  "value",
				"key2": "val2",
			},
		}
		files, err := openFiles([]string{"../../assets/merge/first.yml", "../../assets/merge/second.yml"})
		Expect(err).NotTo(HaveOccurred())
		ev, err := mergeAllDocs(files, mergeOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(ev.Tree).To(Equal(expect))
	})

	It("Succeeds with valid files + json", func() {
		expect := map[interface{}]interface{}{
			"key":           "overridden",
			"array_append":  []interface{}{"one", "two", "three"},
			"array_prepend": []interface{}{"three", "four", "five"},
			"array_replace": []interface{}{[]interface{}{1, 2, 3}},
			"array_inline": []interface{}{
				map[interface{}]interface{}{"name": "first_elem", "val": "overwritten"},
				"second_elem was overwritten",
				"third elem is appended",
			},
			"array_default": []interface{}{
				"FIRST",
				"SECOND",
				"third",
			},
			"array_map_default": []interface{}{
				map[interface{}]interface{}{
					"name": "AAA",
					"k1":   "key 1",
					"k2":   "updated",
				},
				map[interface{}]interface{}{
					"name": "BBB",
					"k2":   "final",
					"k3":   "original",
				},
			},
			"map": map[interface{}]interface{}{
				"key":  "value",
				"key2": "val2",
			},
		}
		files, err := openFiles([]string{"../../assets/merge/first.json", "../../assets/merge/second.yml"})
		Expect(err).NotTo(HaveOccurred())
		ev, err := mergeAllDocs(files, mergeOpts{})
		Expect(err).NotTo(HaveOccurred())
		Expect(ev.Tree).To(Equal(expect))
	})
})
