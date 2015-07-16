package main

import (
	"fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParseYAML(t *testing.T) {
	Convey("parseYAML()", t, func() {
		Convey("returns error for invalid yaml data", func() {
			data := `
asdf: fdsa
- asdf: fdsa
`
			obj, err := parseYAML([]byte(data))
			So(err.Error(), ShouldStartWith, "unmarshal []byte to yaml failed:")
			So(obj, ShouldBeNil)
		})
		Convey("returns error if yaml was not a top level map", func() {
			data := `
- 1
- 2
`
			obj, err := parseYAML([]byte(data))
			So(err.Error(), ShouldStartWith, "Root of YAML document is not a hash/map:")
			So(obj, ShouldBeNil)
		})
		Convey("returns expected datastructure from valid yaml", func() {
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
			So(obj, ShouldResemble, expect)
			So(err, ShouldBeNil)
		})
	})
}

func TestMergeAllDocs(t *testing.T) {
	Convey("mergeAllDocs()", t, func() {
		Convey("Fails with readFile error on bad first doc", func() {
			target := map[interface{}]interface{}{}
			err := mergeAllDocs(target, []string{"assets/merge/nonexistent.yml", "assets/merge/second.yml"})
			So(err.Error(), ShouldStartWith, "Error reading file assets/merge/nonexistent.yml:")
		})
		Convey("Fails with parseYAML error on bad second doc", func() {
			target := map[interface{}]interface{}{}
			err := mergeAllDocs(target, []string{"assets/merge/first.yml", "assets/merge/bad.yml"})
			So(err.Error(), ShouldStartWith, "assets/merge/bad.yml: Root of YAML document is not a hash/map:")
		})
		Convey("Succeeds with valid files + yaml", func() {
			target := map[interface{}]interface{}{}
			expect := map[interface{}]interface{}{
				"key":           "overridden",
				"array_append":  []interface{}{"one", "two", "three"},
				"array_prepend": []interface{}{"three", "four", "five"},
				"array_inline": []interface{}{
					map[interface{}]interface{}{"name": "first_elem", "val": "overwritten"},
					"second_elem was overwritten",
					"third elem is appended",
				},
				"map": map[interface{}]interface{}{
					"key":  "value",
					"key2": "val2",
				},
			}
			err := mergeAllDocs(target, []string{"assets/merge/first.yml", "assets/merge/second.yml"})
			So(err, ShouldBeNil)
			So(target, ShouldResemble, expect)
		})
	})
}

func TestMain(t *testing.T) {
	Convey("main()", t, func() {
		var stdout string
		printfStdOut = func(format string, args ...interface{}) {
			stdout = fmt.Sprintf(format, args...)
		}
		var stderr string
		printfStdErr = func(format string, args ...interface{}) {
			stderr = fmt.Sprintf(format, args...)
		}

		rc := 256 // invalid return code to catch any issues
		exit = func(code int) {
			rc = code
		}

		Convey("Should output usage if no args", func() {
			os.Args = []string{"program_name"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "Usage: "+os.Args[0]+" <first.yml> ... <nth.yml>\n")
			So(rc, ShouldEqual, 1)
		})
		Convey("Should panic on errors merging docs", func() {
			os.Args = []string{"program_name", "assets/merge/bad.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldStartWith, "assets/merge/bad.yml: Root of YAML document is not a hash/map:")
			So(rc, ShouldEqual, 2)
		})
		/* Fixme - how to trigger this?
		Convey("Should panic on errors marshalling yaml", func () {
		})
		*/
		Convey("Should output merged yaml on success", func() {
			os.Args = []string{"program_name", "assets/merge/first.yml", "assets/merge/second.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, `array_append:
- one
- two
- three
array_inline:
- name: first_elem
  val: overwritten
- second_elem was overwritten
- third elem is appended
array_prepend:
- three
- four
- five
key: overridden
map:
  key: value
  key2: val2

`)
		})
	})
}
