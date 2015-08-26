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
		Convey("Fails with mergeMap error", func() {
			target := map[interface{}]interface{}{}
			err := mergeAllDocs(target, []string{"assets/merge/first.yml", "assets/merge/error.yml"})
			So(err.Error(), ShouldStartWith, "assets/merge/error.yml: $.array_inline.0: new object is a string, not a map - cannot merge using keys")
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

		usage = func() {
			stderr = "usage was called"
			exit(1)
		}

		Convey("Should output usage if bad args are passed", func() {
			os.Args = []string{"spruce", "fdsafdada"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "usage was called")
			So(rc, ShouldEqual, 1)
		})
		Convey("Should output usage if no args at all", func() {
			os.Args = []string{"spruce"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "usage was called")
			So(rc, ShouldEqual, 1)
		})
		Convey("Should output usage if no args to merge", func() {
			os.Args = []string{"spruce", "merge"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "usage was called")
			So(rc, ShouldEqual, 1)
		})
		Convey("Should output version", func() {
			Convey("When '-v' is specified", func() {
				os.Args = []string{"spruce", "-v"}
				stdout = ""
				stderr = ""
				main()
				So(stdout, ShouldEqual, "")
				So(stderr, ShouldEqual, fmt.Sprintf("spruce - Version %s\n", VERSION))
				So(rc, ShouldEqual, 0)
			})
			Convey("When '--version' is specified", func() {
				os.Args = []string{"spruce", "--version"}
				stdout = ""
				stderr = ""
				main()
				So(stdout, ShouldEqual, "")
				So(stderr, ShouldEqual, fmt.Sprintf("spruce - Version %s\n", VERSION))
				So(rc, ShouldEqual, 0)
			})
		})
		Convey("Should panic on errors merging docs", func() {
			os.Args = []string{"spruce", "merge", "assets/merge/bad.yml"}
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
			os.Args = []string{"spruce", "merge", "assets/merge/first.yml", "assets/merge/second.yml"}
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
			So(stderr, ShouldEqual, "")
		})
		Convey("Should not fail when handling concourse-style yaml and --concourse", func() {
			os.Args = []string{"spruce", "--concourse", "merge", "assets/concourse/first.yml", "assets/concourse/second.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, `jobs:
- curlies: {{my-variable_123}}
  name: thing1
- curlies: {{more}}
  name: thing2

`)
			So(stderr, ShouldEqual, "")
		})

		Convey("Should handle pruning", func() {
			os.Args = []string{"spruce", "merge", "assets/prune/first.yml", "assets/prune/second.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, `level2:
  level3:
    retained: yea
  retained: yea
retained: yea

`)

			So(stderr, ShouldEqual, "")
		})

		Convey("Should handle de-referencing", func() {
			os.Args = []string{"spruce", "merge", "assets/dereference/first.yml", "assets/dereference/second.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, `jobs:
- name: my-server
  static_ips:
  - 192.168.1.0
properties:
  client:
    servers:
    - 192.168.1.0

`)
			So(stderr, ShouldEqual, "")
		})
		Convey("De-referencing shouldn't result in cyclical data structures", func() {
			os.Args = []string{"spruce", "merge", "assets/dereference/cyclic-data.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, `people:
  anne:
    givenName: Anne
    spouse:
      givenName: Bartholomew
      spouse: (( people.anne ))
      ssn: 456789
      surName: Jennings
    ssn: 123456
    surName: Bolswenn
  bart:
    givenName: Bartholomew
    spouse:
      givenName: Anne
      spouse:
        givenName: Bartholomew
        spouse: (( people.anne ))
        ssn: 456789
        surName: Jennings
      ssn: 123456
      surName: Bolswenn
    ssn: 456789
    surName: Jennings

`)
			So(stderr, ShouldEqual, "")
		})
		Convey("Should output error on bad de-reference", func() {
			os.Args = []string{"spruce", "merge", "assets/dereference/bad.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldStartWith, "$.bad.dereference: Unable to resolve `(( my.value ))`")
			So(rc, ShouldEqual, 2)
		})
	})
}

func TestDebug(t *testing.T) {
	var stderr string
	usage = func() {}
	printfStdErr = func(format string, args ...interface{}) {
		stderr = fmt.Sprintf(format, args...)
	}
	Convey("debug", t, func() {
		Convey("Outputs when debug is set to true", func() {
			stderr = ""
			debug = true
			DEBUG("test debugging")
			So(stderr, ShouldEqual, "DEBUG> test debugging\n")
		})
		Convey("Multi-line debug inputs are each prefixed", func() {
			stderr = ""
			debug = true
			DEBUG("test debugging\nsecond line")
			So(stderr, ShouldEqual, "DEBUG> test debugging\nDEBUG> second line\n")
		})
		Convey("Doesn't output when debug is set to false", func() {
			stderr = ""
			debug = false
			DEBUG("test debugging")
			So(stderr, ShouldEqual, "")
		})
	})
	Convey("debug flags:", t, func() {
		Convey("-D enables debugging", func() {
			os.Args = []string{"spruce", "-D"}
			debug = false
			main()
			So(debug, ShouldBeTrue)
		})
		Convey("--debug enables debugging", func() {
			os.Args = []string{"spruce", "--debug"}
			debug = false
			main()
			So(debug, ShouldBeTrue)
		})
		Convey("DEBUG=\"tRuE\" enables debugging", func() {
			os.Setenv("DEBUG", "tRuE")
			os.Args = []string{"spruce"}
			debug = false
			main()
			So(debug, ShouldBeTrue)
		})
		Convey("DEBUG=1 enables debugging", func() {
			os.Setenv("DEBUG", "1")
			os.Args = []string{"spruce"}
			debug = false
			main()
			So(debug, ShouldBeTrue)
		})
		Convey("DEBUG=randomval enables debugging", func() {
			os.Setenv("DEBUG", "randomval")
			os.Args = []string{"spruce"}
			debug = false
			main()
			So(debug, ShouldBeTrue)
		})
		Convey("DEBUG=\"fAlSe\" disables debugging", func() {
			os.Setenv("DEBUG", "fAlSe")
			os.Args = []string{"spruce"}
			debug = false
			main()
			So(debug, ShouldBeFalse)
		})
		Convey("DEBUG=0 disables debugging", func() {
			os.Setenv("DEBUG", "0")
			os.Args = []string{"spruce"}
			debug = false
			main()
			So(debug, ShouldBeFalse)
		})
		Convey("DEBUG=\"\" disables debugging", func() {
			os.Setenv("DEBUG", "")
			os.Args = []string{"spruce"}
			debug = false
			main()
			So(debug, ShouldBeFalse)
		})
	})
}

func TestQuoteConcourse(t *testing.T) {
	Convey("quoteConcourse()", t, func() {
		Convey("Correctly double-quotes incoming {{\\S}} patterns", func() {
			Convey("adds quotes", func() {
				input := []byte("name: {{var-_1able}}")
				So(string(quoteConcourse(input)), ShouldEqual, "name: \"{{var-_1able}}\"")
			})
		})
		Convey("doesn't affect regularly quoted things", func() {
			input := []byte("name: \"my value\"")
			So(string(quoteConcourse(input)), ShouldEqual, "name: \"my value\"")
		})
	})
}
func TestDequoteConcourse(t *testing.T) {
	Convey("dequoteConcourse()", t, func() {
		Convey("Correctly removes quotes from incoming {{\\S}} patterns", func() {
			Convey("with single quotes", func() {
				input := []byte("name: '{{var-_1able}}'")
				So(dequoteConcourse(input), ShouldEqual, "name: {{var-_1able}}")
			})
			Convey("with double quotes", func() {
				input := []byte("name: \"{{var-_1able}}\"")
				So(dequoteConcourse(input), ShouldEqual, "name: {{var-_1able}}")
			})
		})
		Convey("doesn't affect regularly quoted things", func() {
			input := []byte("name: \"my value\"")
			So(dequoteConcourse(input), ShouldEqual, "name: \"my value\"")
		})
	})
}
