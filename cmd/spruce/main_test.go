package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	// Use geofffranks forks to persist the fix in https://github.com/go-yaml/yaml/pull/133/commits
	// Also https://github.com/go-yaml/yaml/pull/195
	"github.com/geofffranks/simpleyaml"
	"github.com/geofffranks/yaml"

	. "github.com/geofffranks/spruce/log"
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
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "unmarshal []byte to yaml failed:")
			So(obj, ShouldBeNil)
		})
		Convey("returns error if yaml was not a top level map", func() {
			data := `
- 1
- 2
`
			obj, err := parseYAML([]byte(data))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Root of YAML document is not a hash/map:")
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
			err := mergeAllDocs(target, []string{"../../assets/merge/nonexistent.yml", "../../assets/merge/second.yml"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Error reading file ../../assets/merge/nonexistent.yml:")
		})
		Convey("Fails with parseYAML error on bad second doc", func() {
			target := map[interface{}]interface{}{}
			err := mergeAllDocs(target, []string{"../../assets/merge/first.yml", "../../assets/merge/bad.yml"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "../../assets/merge/bad.yml: Root of YAML document is not a hash/map:")
		})
		Convey("Fails with mergeMap error", func() {
			target := map[interface{}]interface{}{}
			err := mergeAllDocs(target, []string{"../../assets/merge/first.yml", "../../assets/merge/error.yml"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "$.array_inline.0: new object is a string, not a map - cannot merge using keys")
		})
		Convey("Succeeds with valid files + yaml", func() {
			target := map[interface{}]interface{}{}
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
			err := mergeAllDocs(target, []string{"../../assets/merge/first.yml", "../../assets/merge/second.yml"})
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
		Convey("Should error if no args to merge and no files listed", func() {
			os.Args = []string{"spruce", "merge"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "Error reading STDIN: no data found. Did you forget to pipe data to STDIN, or specify yaml files to merge?\n")
			So(rc, ShouldEqual, 2)
		})
		Convey("Should output version", func() {
			Convey("When '-v' is specified", func() {
				os.Args = []string{"spruce", "-v"}
				stdout = ""
				stderr = ""
				main()
				So(stdout, ShouldStartWith, fmt.Sprintf("spruce - Version %s", Version))
				So(stderr, ShouldEqual, "")
				So(rc, ShouldEqual, 0)
			})
			Convey("When '--version' is specified", func() {
				os.Args = []string{"spruce", "--version"}
				stdout = ""
				stderr = ""
				main()
				So(stdout, ShouldStartWith, fmt.Sprintf("spruce - Version %s", Version))
				So(stderr, ShouldEqual, "")
				So(rc, ShouldEqual, 0)
			})
		})
		Convey("Should panic on errors merging docs", func() {
			os.Args = []string{"spruce", "merge", "../../assets/merge/bad.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldContainSubstring, "../../assets/merge/bad.yml: Root of YAML document is not a hash/map:")
			So(rc, ShouldEqual, 2)
		})
		/* Fixme - how to trigger this?
		Convey("Should panic on errors marshalling yaml", func () {
		})
		*/
		Convey("Should output merged yaml on success", func() {
			os.Args = []string{"spruce", "merge", "../../assets/merge/first.yml", "../../assets/merge/second.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, `array_append:
- one
- two
- three
array_default:
- FIRST
- SECOND
- third
array_inline:
- name: first_elem
  val: overwritten
- second_elem was overwritten
- third elem is appended
array_map_default:
- k1: key 1
  k2: updated
  name: AAA
- k2: final
  k3: original
  name: BBB
array_prepend:
- three
- four
- five
array_replace:
- - 1
  - 2
  - 3
key: overridden
map:
  key: value
  key2: val2

`)
			So(stderr, ShouldEqual, "")
		})
		Convey("Should not fail when handling concourse-style yaml and --concourse", func() {
			os.Args = []string{"spruce", "--concourse", "merge", "../../assets/concourse/first.yml", "../../assets/concourse/second.yml"}
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

		Convey("Should handle de-referencing", func() {
			os.Args = []string{"spruce", "merge", "../../assets/dereference/first.yml", "../../assets/dereference/second.yml"}
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
		Convey("De-referencing cyclical datastructures should throw an error", func() {
			os.Args = []string{"spruce", "merge", "../../assets/dereference/cyclic-data.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, "")
			So(stderr, ShouldContainSubstring, "max recursion depth. You seem to have a self-referencing dataset\n")
			So(rc, ShouldEqual, 2)
		})
		Convey("Dereferencing multiple values should behave as desired", func() {
			//UsedIPs = map[string]string{} // required because of globalness
			os.Args = []string{"spruce", "merge", "../../assets/dereference/multi-value.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `jobs:
- instances: 1
  name: api_z1
  networks:
  - name: net1
    static_ips:
    - 192.168.1.2
- instances: 1
  name: api_z2
  networks:
  - name: net2
    static_ips:
    - 192.168.2.2
networks:
- name: net1
  subnets:
  - cloud_properties: random
    static:
    - 192.168.1.2 - 192.168.1.30
- name: net2
  subnets:
  - cloud_properties: random
    static:
    - 192.168.2.2 - 192.168.2.30
properties:
  api_server_primary: 192.168.1.2
  api_servers:
  - 192.168.1.2
  - 192.168.2.2

`)
		})
		Convey("Should output error on bad de-reference", func() {
			os.Args = []string{"spruce", "merge", "../../assets/dereference/bad.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldContainSubstring, "$.bad.dereference: Unable to resolve `my.value`")
			So(rc, ShouldEqual, 2)
		})
		Convey("Pruning should happen after de-referencing", func() {
			os.Args = []string{"spruce", "merge", "--prune", "jobs", "--prune", "properties.client.servers", "../../assets/dereference/first.yml", "../../assets/dereference/second.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `properties:
  client: {}

`)
		})
		Convey("can dereference ~ / null values", func() {
			os.Args = []string{"spruce", "merge", "--prune", "meta", "../../assets/dereference/null.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `value: null

`)
		})
		Convey("can dereference nestedly", func() {
			os.Args = []string{"spruce", "merge", "../../assets/dereference/multi.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `name1: name
name2: name
name3: name
name4: name

`)
		})
		Convey("static_ips() failures return errors to the user", func() {
			os.Args = []string{"spruce", "merge", "../../assets/static_ips/jobs.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldContainSubstring, ".static_ips: `$.networks` could not be found in the datastructure\n")
			So(stdout, ShouldEqual, "")
		})
		Convey("static_ips() get resolved, and are resolved prior to dereferencing", func() {
			os.Args = []string{"spruce", "merge", "../../assets/static_ips/properties.yml", "../../assets/static_ips/jobs.yml", "../../assets/static_ips/network.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `jobs:
- instances: 3
  name: api_z1
  networks:
  - name: net1
    static_ips:
    - 10.0.0.2
    - 10.0.0.3
    - 10.0.0.4
networks:
- name: net1
  subnets:
  - static:
    - 10.0.0.2 - 10.0.0.20
properties:
  api_servers:
  - 10.0.0.2
  - 10.0.0.3
  - 10.0.0.4

`)
		})
		Convey("Included yaml file is escaped", func() {
			os.Setenv("SPRUCE_FILE_BASE_PATH", "../../assets/file_operator")
			os.Args = []string{"spruce", "merge", "../../assets/file_operator/test.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, `content:
  meta_test:
    stuff: |
      ---
      meta:
        filename: test.yml

      content:
        meta_test:
          stuff: (( file meta.filename ))
meta:
  filename: test.yml

`)
			So(stderr, ShouldEqual, "")
		})

		Convey("Parameters override their requirement", func() {
			os.Args = []string{"spruce", "merge", "../../assets/params/global.yml", "../../assets/params/good.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, `cpu: 3
nested:
  key:
    override: true
networks:
- true
storage: 4096

`)
			So(stderr, ShouldEqual, "")
		})
		Convey("Parameters must be specified", func() {
			os.Args = []string{"spruce", "merge", "../../assets/params/global.yml", "../../assets/params/fail.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, "")
			So(stderr, ShouldContainSubstring, "$.nested.key.override: provide nested override\n")
		})
		Convey("Pruning takes place after parameters", func() {
			os.Args = []string{"spruce", "merge", "--prune", "nested", "../../assets/params/global.yml", "../../assets/params/fail.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, `1 error(s) detected:
 - $.nested.key.override: provide nested override


`)
			So(stdout, ShouldEqual, "")
		})
		Convey("string concatenation works", func() {
			os.Args = []string{"spruce", "merge", "--prune", "local", "--prune", "env", "--prune", "cluster", "../../assets/concat/concat.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `ident: c=mjolnir/prod;1234567890-abcdef

`)
		})
		Convey("string concatenation handles non-strings correctly", func() {
			os.Args = []string{"spruce", "merge", "--prune", "local", "../../assets/concat/coerce.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `url: http://domain.example.com/?v=1.3&rev=42

`)
		})
		Convey("string concatenation failure detected", func() {
			os.Args = []string{"spruce", "merge", "../../assets/concat/fail.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, "")
			So(stderr, ShouldContainSubstring, "$.ident: Unable to resolve `local.sites.42.uuid`:")
		})
		Convey("string concatentation handles multiple levels of reference", func() {
			os.Args = []string{"spruce", "merge", "../../assets/concat/multi.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `bar: quux.bar
baz: quux.bar.baz
foo: quux.bar.baz.foo
quux: quux

`)
			Convey("string concatenation handles infinite loop self-reference", func() {
				os.Args = []string{"spruce", "merge", "../../assets/concat/loop.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stdout, ShouldEqual, "")
				So(stderr, ShouldContainSubstring, "cycle detected")
			})
		})

		Convey("only param errors are displayed, if present", func() {
			os.Args = []string{"spruce", "merge", "../../assets/errors/multi.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, "")
			So(stderr, ShouldEqual, ""+
				"1 error(s) detected:\n"+
				" - $.an-error: missing param!\n"+
				"\n\n"+
				"")
		})

		Convey("multiple errors of the same type on the same level are displayed", func() {
			os.Args = []string{"spruce", "merge", "../../assets/errors/multi2.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, "")
			So(stderr, ShouldEqual, ""+
				"3 error(s) detected:\n"+
				" - $.a: first\n"+
				" - $.b: second\n"+
				" - $.c: third\n"+
				"\n\n"+
				"")
		})

		Convey("json command converts YAML to JSON", func() {
			os.Args = []string{"spruce", "json", "../../assets/json/in.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `{"map":{"list":["string",42,{"map":"of things"}]}}`+"\n")
		})

		Convey("json command handles malformed YAML", func() {
			os.Args = []string{"spruce", "json", "../../assets/json/malformed.yml"}
			stdout = ""
			stderr = ""
			main()
			So(stdout, ShouldEqual, "")
			So(stderr, ShouldContainSubstring, "Root of YAML document is not a hash/map:")
		})

		Convey("Adding (dynamic) prune support for list entries (edge case scenario)", func() {
			os.Args = []string{"spruce", "merge", "../../assets/prune/prune-in-lists/fileA.yml", "../../assets/prune/prune-in-lists/fileB.yml"}
			stdout = ""
			stderr = ""

			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `meta:
  list:
  - one
  - three

`)
		})

		Convey("Adding (static) prune support for list entries (edge case scenario)", func() {
			os.Args = []string{"spruce", "merge", "--prune", "meta.list.1", "../../assets/prune/prune-in-lists/fileA.yml"}
			stdout = ""
			stderr = ""

			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `meta:
  list:
  - one
  - three

`)
		})

		Convey("Issue - prune and inject cause side-effect", func() {
			os.Args = []string{"spruce", "merge", "--prune", "meta", "../../assets/prune/prune-issue-with-inject/fileA.yml", "../../assets/prune/prune-issue-with-inject/fileB.yml"}
			stdout = ""
			stderr = ""

			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `jobs:
- instances: 2
  name: main-job
  templates:
  - name: one
  - name: two
  update:
    canaries: 1
    max_in_flight: 3
- instances: 1
  name: another-job
  templates:
  - name: one
  - name: two
  update:
    canaries: 2

`)
		})

		Convey("Issue - prune and new-list-entry cause side-effect", func() {
			os.Args = []string{"spruce", "merge", "--prune", "meta", "../../assets/prune/prune-issue-in-lists-with-new-entry/fileA.yml", "../../assets/prune/prune-issue-in-lists-with-new-entry/fileB.yml"}
			stdout = ""
			stderr = ""

			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `list:
- name: A
  release: A
  version: A
- name: B
  release: B
  version: B
- name: C
  release: C
  version: C
- name: D
  release: D

`)
		})

		Convey("Issue #158 prune doesn't work when goes at the end (regression?) - variant A (https://github.com/geofffranks/spruce/issues/158)", func() {
			os.Args = []string{"spruce", "merge", "../../assets/prune/issue-158/test.yml", "../../assets/prune/issue-158/prune.yml"}
			stdout = ""
			stderr = ""

			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `test1: t2

`)
		})

		Convey("Issue #158 prune doesn't work when goes at the end (regression?) - variant B (https://github.com/geofffranks/spruce/issues/158)", func() {
			os.Args = []string{"spruce", "merge", "../../assets/prune/issue-158/prune.yml", "../../assets/prune/issue-158/test.yml"}
			stdout = ""
			stderr = ""

			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `test1: t2

`)
		})

		Convey("Issue #156 Can use concat with static ips", func() {
			os.Args = []string{"spruce", "merge", "../../assets/static_ips/issue-156/concat.yml"}
			stdout = ""
			stderr = ""

			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `jobs:
- instances: 1
  name: pepe
  networks:
  - name: cf1
    static_ips:
    - 10.4.5.4
meta:
  network_prefix: "10.4"
networks:
- name: cf1
  subnets:
  - range: 10.4.36.0/24
    static:
    - 10.4.5.4 - 10.4.5.100

`)
		})

		Convey("Issue #194 Globs with missing sub-items track data flow deps properly", func() {
			os.Args = []string{"spruce", "merge", "../../assets/static_ips/vips-plus-grab.yml"}
			stdout = ""
			stderr = ""

			main()
			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, `jobs:
- instances: 1
  name: bosh
  networks:
  - name: stuff
    static_ips:
    - 1.2.3.4
meta:
  ips:
  - 1.2.3.4
networks:
- name: stuff
  subnets:
  - static:
    - 1.2.3.4
- name: stuff2
  type: vip

`)
		})

		Convey("Empty operator works", func() {

			var baseFile, mergeFile string
			baseFile = "../../assets/empty/base.yml"

			testEmpty := func(files ...string) {
				os.Args = append([]string{"spruce", "merge"}, files...)
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `meta:
  first: {}
  second: []
  third: ""

`)
			}

			Convey("when merging over maps", func() {

				Convey("with references as the type", func() {
					mergeFile = "../../assets/empty/references.yml"
					testEmpty(baseFile, mergeFile)
				})
				Convey("with literals as the type", func() {
					mergeFile = "../../assets/empty/literals.yml"
					testEmpty(baseFile, mergeFile)
				})
			})

			Convey("when merging over nothing", func() {
				Convey("with references as the type", func() {
					mergeFile = "../../assets/empty/references.yml"
					testEmpty(mergeFile)
				})
				Convey("with literals as the type", func() {
					mergeFile = "../../assets/empty/literals.yml"
					testEmpty(mergeFile)
				})
			})
		})

		Convey("Join operator works", func() {
			Convey("when dependencies could cause improper evaluation order", func() {
				os.Args = []string{"spruce", "merge", "../../assets/join/issue-155/deps.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `b:
- hello
- world
greeting: hello
output:
- hello world
- hello bye
z:
- hello
- bye

`)
			})
		})

		Convey("Calc operator works", func() {
			Convey("Calc comes with built-in functions", func() {
				os.Args = []string{"spruce", "merge", "--prune", "meta", "../../assets/calc/functions.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `properties:
  homework:
    ceil: 9
    floor: 3
    max: 8.333
    min: 3.666
    mod: 1.001
    pow: 2374.9685
    sqrt: 2.8866937

`)
			})

			Convey("Calc works with dependencies", func() {
				os.Args = []string{"spruce", "merge", "--prune", "meta", "../../assets/calc/dependencies.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `jobs:
- instances: 4
  name: big_ones
- instances: 1
  name: small_ones
- instances: 2
  name: extra_ones

`)
			})

			Convey("Calc expects only one argument which is a quoted mathematical expression (as a Literal in Spruce)", func() {
				os.Args = []string{"spruce", "merge", "--prune", "meta", "../../assets/calc/wrong-syntax.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, `2 error(s) detected:
 - $.jobs.one.instances: calc operator only expects one argument containing the expression
 - $.jobs.two.instances: calc operator argument is suppose to be a quoted mathematical expression (type Literal)


`)
				So(stdout, ShouldEqual, "")
			})

			Convey("Calc operator does not support named variables", func() {
				os.Args = []string{"spruce", "merge", "--prune", "meta", "../../assets/calc/no-named-variables.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, `1 error(s) detected:
 - $.jobs.one.instances: calc operator does not support named variables in expression: pi, r


`)
				So(stdout, ShouldEqual, "")
			})

			Convey("Calc operator checks input for built-in functions", func() {
				os.Args = []string{"spruce", "merge", "--prune", "meta", "../../assets/calc/bad-functions.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, `7 error(s) detected:
 - $.properties.homework.ceil: ceil function expects one argument of type float64
 - $.properties.homework.floor: floor function expects one argument of type float64
 - $.properties.homework.max: max function expects two arguments of type float64
 - $.properties.homework.min: min function expects two arguments of type float64
 - $.properties.homework.mod: mod function expects two arguments of type float64
 - $.properties.homework.pow: pow function expects two arguments of type float64
 - $.properties.homework.sqrt: sqrt function expects one argument of type float64


`)
				So(stdout, ShouldEqual, "")
			})

			Convey("Calc operator checks referenced types", func() {
				os.Args = []string{"spruce", "merge", "--prune", "meta", "../../assets/calc/wrong-type.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, `4 error(s) detected:
 - $.properties.homework.list: path meta.list is of type slice, which cannot be used in calculations
 - $.properties.homework.map: path meta.map is of type map, which cannot be used in calculations
 - $.properties.homework.nil: path meta.nil references a nil value, which cannot be used in calculations
 - $.properties.homework.string: path meta.string is of type string, which cannot be used in calculations


`)
				So(stdout, ShouldEqual, "")
			})
		})

		Convey("YAML output is ordered the same way each time (#184)", func() {
			for i := 0; i < 30; i++ {
				os.Args = []string{"spruce", "merge", "../../assets/output-order/sample.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `properties:
  cc:
    quota_definitions:
      q2GB:
        non_basic_services_allowed: true
      q4GB:
        non_basic_services_allowed: true
      q256MB:
        non_basic_services_allowed: true

`)
			}
		})

		Convey("Cherry picking test cases", func() {
			Convey("Cherry pick just one root level path", func() {
				os.Args = []string{"spruce", "merge", "--cherry-pick", "properties", "../../assets/cherry-pick/fileA.yml", "../../assets/cherry-pick/fileB.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `properties:
  hahn:
    flags: open
    id: b503e54a-c872-4643-a09c-5480c5940d0c
  vb:
    flags: auth,block,read-only
    id: 74a03820-3f81-45ca-afd5-d7d57b947ff1

`)
			})

			Convey("Cherry pick a path that is a list entry", func() {
				os.Args = []string{"spruce", "merge", "--cherry-pick", "releases.vb", "../../assets/cherry-pick/fileA.yml", "../../assets/cherry-pick/fileB.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `releases:
- name: vb

`)
			})

			Convey("Cherry pick a path that is deep down the structure", func() {
				os.Args = []string{"spruce", "merge", "--cherry-pick", "meta.some.deep.structure.maplist", "../../assets/cherry-pick/fileA.yml", "../../assets/cherry-pick/fileB.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `meta:
  some:
    deep:
      structure:
        maplist:
          keyA: valueA
          keyB: valueB

`)
			})

			Convey("Cherry pick a series of different paths at the same time", func() {
				os.Args = []string{"spruce", "merge", "--cherry-pick", "properties", "--cherry-pick", "releases.vb", "--cherry-pick", "meta.some.deep.structure.maplist", "../../assets/cherry-pick/fileA.yml", "../../assets/cherry-pick/fileB.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `meta:
  some:
    deep:
      structure:
        maplist:
          keyA: valueA
          keyB: valueB
properties:
  hahn:
    flags: open
    id: b503e54a-c872-4643-a09c-5480c5940d0c
  vb:
    flags: auth,block,read-only
    id: 74a03820-3f81-45ca-afd5-d7d57b947ff1
releases:
- name: vb

`)
			})

			Convey("Cherry pick a path and prune something at the same time in a map", func() {
				os.Args = []string{"spruce", "merge", "--cherry-pick", "properties", "--prune", "properties.vb.flags", "../../assets/cherry-pick/fileA.yml", "../../assets/cherry-pick/fileB.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `properties:
  hahn:
    flags: open
    id: b503e54a-c872-4643-a09c-5480c5940d0c
  vb:
    id: 74a03820-3f81-45ca-afd5-d7d57b947ff1

`)
			})

			Convey("Cherry picking should fail if you cherry-pick a prune path", func() {
				os.Args = []string{"spruce", "merge", "--cherry-pick", "properties", "--prune", "properties", "../../assets/cherry-pick/fileA.yml", "../../assets/cherry-pick/fileB.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "1 error(s) detected:\n"+
					" - `$.properties` could not be found in the datastructure\n\n\n")
				So(stdout, ShouldEqual, "")
			})

			Convey("Cherry picking should fail if picking a sub-level path while prune wipes the parent", func() {
				os.Args = []string{"spruce", "merge", "--cherry-pick", "releases.vb", "--prune", "releases", "../../assets/cherry-pick/fileA.yml", "../../assets/cherry-pick/fileB.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "1 error(s) detected:\n"+
					" - `$.releases` could not be found in the datastructure\n\n\n")
				So(stdout, ShouldEqual, "")
			})

			Convey("Cherry pick a list entry path of a list that uses 'key' as its identifier", func() {
				os.Args = []string{"spruce", "merge", "--cherry-pick", "list.two", "../../assets/cherry-pick/key-based-list.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `list:
- desc: The second one
  key: two
  version: v2

`)
			})

			Convey("Cherry pick a list entry path of a list that uses 'id' as its identifier", func() {
				os.Args = []string{"spruce", "merge", "--cherry-pick", "list.two", "../../assets/cherry-pick/id-based-list.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `list:
- desc: The second one
  id: two
  version: v2

`)
			})

			Convey("Cherry pick one list entry path that references the index", func() {
				os.Args = []string{"spruce", "merge", "--cherry-pick", "list.1", "../../assets/cherry-pick/name-based-list.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `list:
- desc: The second one
  name: two
  version: v2

`)
			})

			Convey("Cherry pick two list entry paths that reference indexes", func() {
				os.Args = []string{"spruce", "merge", "--cherry-pick", "list.1", "--cherry-pick", "list.4", "../../assets/cherry-pick/name-based-list.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `list:
- desc: The fifth one
  name: five
  version: v5
- desc: The second one
  name: two
  version: v2

`)
			})

			Convey("Cherry pick one list entry path that references an invalid index", func() {
				os.Args = []string{"spruce", "merge", "--cherry-pick", "list.10", "../../assets/cherry-pick/name-based-list.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "1 error(s) detected:\n"+
					" - `$.list.10` could not be found in the datastructure\n\n\n")
				So(stdout, ShouldEqual, "")
			})

			Convey("Cherry pick should only pick the exact name based on the path", func() {
				os.Args = []string{"spruce", "merge", "--cherry-pick", "map", "--prune", "subkey", "../../assets/cherry-pick/test-exact-names.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `map:
  other: value
  subkey: this is the real subkey

`)
			})

			Convey("Cherry pick should only evaluate the dynamic operators that are relevant", func() {
				os.Args = []string{"spruce", "merge", "--cherry-pick", "params", "../../assets/cherry-pick/partial-eval.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `params:
  mode: default
  name: sandbox-thing

`)
			})
		})

		Convey("non-specific node tags specific test cases", func() {
			Convey("non-specific node tags test case - style 1", func() {
				os.Args = []string{"spruce", "merge", "../../assets/non-specific-node-tags-issue/fileA-1.yml", "../../assets/non-specific-node-tags-issue/fileB.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `some:
  yaml:
    structure:
      certificate: |
        -----BEGIN CERTIFICATE-----
        QSBzcHJ1Y2UgaXMgYSB0cmVlIG9mIHRoZSBnZW51cyBQaWNlYSAvcGHJqsuIc2nL
        kMmZLyxbMV0gYSBnZW51cyBvZiBhYm91dCAzNSBzcGVjaWVzIG9mIGNvbmlmZXJv
        dXMgZXZlcmdyZWVuIHRyZWVzIGluIHRoZSBGYW1pbHkgUGluYWNlYWUsIGZvdW5k
        IGluIHRoZSBub3J0aGVybiB0ZW1wZXJhdGUgYW5kIGJvcmVhbCAodGFpZ2EpIHJl
        Z2lvbnMgb2YgdGhlIGVhcnRoLiBTcHJ1Y2VzIGFyZSBsYXJnZSB0cmVlcywgZnJv
        bSBhYm91dCAyMOKAkzYwIG1ldHJlcyAoYWJvdXQgNjDigJMyMDAgZmVldCkgdGFs
        bCB3aGVuIG1hdHVyZSwgYW5kIGNhbiBiZSBkaXN0aW5ndWlzaGVkIGJ5IHRoZWly
        IHdob3JsZWQgYnJhbmNoZXMgYW5kIGNvbmljYWwgZm9ybS4gVGhlIG5lZWRsZXMs
        IG9yIGxlYXZlcywgb2Ygc3BydWNlIHRyZWVzIGFyZSBhdHRhY2hlZCBzaW5nbHkg
        dG8gdGhlIGJyYW5jaGVzIGluIGEgc3BpcmFsIGZhc2hpb24sIGVhY2ggbmVlZGxl
        IG9uIGEgc21hbGwgcGVnLWxpa2Ugc3RydWN0dXJlLiBUaGUgbmVlZGxlcyBhcmUg
        c2hlZCB3aGVuIDTigJMxMCB5ZWFycyBvbGQsIGxlYXZpbmcgdGhlIGJyYW5jaGVz
        IHJvdWdoIHdpdGggdGhlIHJldGFpbmVkIHBlZ3MgKGFuIGVhc3kgbWVhbnMgb2Yg
        ZGlzdGluZ3Vpc2hpbmcgdGhlbSBmcm9tIG90aGVyIHNpbWlsYXIgZ2VuZXJhLCB3
        aGVyZSB0aGUgYnJhbmNoZXMgYXJlIGZhaXJseSBzbW9vdGgpLgoKU3BydWNlcyBh
        cmUgdXNlZCBhcyBmb29kIHBsYW50cyBieSB0aGUgbGFydmFlIG9mIHNvbWUgTGVw
        aWRvcHRlcmEgKG1vdGggYW5kIGJ1dHRlcmZseSkgc3BlY2llczsgc2VlIGxpc3Qg
        b2YgTGVwaWRvcHRlcmEgdGhhdCBmZWVkIG9uIHNwcnVjZXMuIFRoZXkgYXJlIGFs
        c28gdXNlZCBieSB0aGUgbGFydmFlIG9mIGdhbGwgYWRlbGdpZHMgKEFkZWxnZXMg
        c3BlY2llcykuCgpJbiB0aGUgbW91bnRhaW5zIG9mIHdlc3Rlcm4gU3dlZGVuIHNj
        aWVudGlzdHMgaGF2ZSBmb3VuZCBhIE5vcndheSBzcHJ1Y2UgdHJlZSwgbmlja25h
        bWVkIE9sZCBUamlra28sIHdoaWNoIGJ5IHJlcHJvZHVjaW5nIHRocm91Z2ggbGF5
        ZXJpbmcgaGFzIHJlYWNoZWQgYW4gYWdlIG9mIDksNTUwIHllYXJzIGFuZCBpcyBj
        bGFpbWVkIHRvIGJlIHRoZSB3b3JsZCdzIG9sZGVzdCBrbm93biBsaXZpbmcgdHJl
        ZS4K
        -----END CERTIFICATE-----
      someotherkey: value

`)
			})

			Convey("non-specific node tags test case - style 2", func() {
				os.Args = []string{"spruce", "merge", "../../assets/non-specific-node-tags-issue/fileA-2.yml", "../../assets/non-specific-node-tags-issue/fileB.yml"}
				stdout = ""
				stderr = ""
				main()
				So(stderr, ShouldEqual, "")
				So(stdout, ShouldEqual, `some:
  yaml:
    structure:
      certificate: '-----BEGIN CERTIFICATE----- QSBzcHJ1Y2UgaXMgYSB0cmVlIG9mIHRoZSBnZW51cyBQaWNlYSAvcGHJqsuIc2nL
        kMmZLyxbMV0gYSBnZW51cyBvZiBhYm91dCAzNSBzcGVjaWVzIG9mIGNvbmlmZXJv dXMgZXZlcmdyZWVuIHRyZWVzIGluIHRoZSBGYW1pbHkgUGluYWNlYWUsIGZvdW5k
        IGluIHRoZSBub3J0aGVybiB0ZW1wZXJhdGUgYW5kIGJvcmVhbCAodGFpZ2EpIHJl Z2lvbnMgb2YgdGhlIGVhcnRoLiBTcHJ1Y2VzIGFyZSBsYXJnZSB0cmVlcywgZnJv
        bSBhYm91dCAyMOKAkzYwIG1ldHJlcyAoYWJvdXQgNjDigJMyMDAgZmVldCkgdGFs bCB3aGVuIG1hdHVyZSwgYW5kIGNhbiBiZSBkaXN0aW5ndWlzaGVkIGJ5IHRoZWly
        IHdob3JsZWQgYnJhbmNoZXMgYW5kIGNvbmljYWwgZm9ybS4gVGhlIG5lZWRsZXMs IG9yIGxlYXZlcywgb2Ygc3BydWNlIHRyZWVzIGFyZSBhdHRhY2hlZCBzaW5nbHkg
        dG8gdGhlIGJyYW5jaGVzIGluIGEgc3BpcmFsIGZhc2hpb24sIGVhY2ggbmVlZGxl IG9uIGEgc21hbGwgcGVnLWxpa2Ugc3RydWN0dXJlLiBUaGUgbmVlZGxlcyBhcmUg
        c2hlZCB3aGVuIDTigJMxMCB5ZWFycyBvbGQsIGxlYXZpbmcgdGhlIGJyYW5jaGVz IHJvdWdoIHdpdGggdGhlIHJldGFpbmVkIHBlZ3MgKGFuIGVhc3kgbWVhbnMgb2Yg
        ZGlzdGluZ3Vpc2hpbmcgdGhlbSBmcm9tIG90aGVyIHNpbWlsYXIgZ2VuZXJhLCB3 aGVyZSB0aGUgYnJhbmNoZXMgYXJlIGZhaXJseSBzbW9vdGgpLgoKU3BydWNlcyBh
        cmUgdXNlZCBhcyBmb29kIHBsYW50cyBieSB0aGUgbGFydmFlIG9mIHNvbWUgTGVw aWRvcHRlcmEgKG1vdGggYW5kIGJ1dHRlcmZseSkgc3BlY2llczsgc2VlIGxpc3Qg
        b2YgTGVwaWRvcHRlcmEgdGhhdCBmZWVkIG9uIHNwcnVjZXMuIFRoZXkgYXJlIGFs c28gdXNlZCBieSB0aGUgbGFydmFlIG9mIGdhbGwgYWRlbGdpZHMgKEFkZWxnZXMg
        c3BlY2llcykuCgpJbiB0aGUgbW91bnRhaW5zIG9mIHdlc3Rlcm4gU3dlZGVuIHNj aWVudGlzdHMgaGF2ZSBmb3VuZCBhIE5vcndheSBzcHJ1Y2UgdHJlZSwgbmlja25h
        bWVkIE9sZCBUamlra28sIHdoaWNoIGJ5IHJlcHJvZHVjaW5nIHRocm91Z2ggbGF5 ZXJpbmcgaGFzIHJlYWNoZWQgYW4gYWdlIG9mIDksNTUwIHllYXJzIGFuZCBpcyBj
        bGFpbWVkIHRvIGJlIHRoZSB3b3JsZCdzIG9sZGVzdCBrbm93biBsaXZpbmcgdHJl ZS4K -----END
        CERTIFICATE-----'
      someotherkey: value

`)
			})
		})
	})
}

func TestDebug(t *testing.T) {
	usage = func() {}
	Convey("debug flags:", t, func() {
		Convey("-D enables debugging", func() {
			os.Args = []string{"spruce", "-D"}
			DebugOn = false
			main()
			So(DebugOn, ShouldBeTrue)
		})
		Convey("--debug enables debugging", func() {
			os.Args = []string{"spruce", "--debug"}
			DebugOn = false
			main()
			So(DebugOn, ShouldBeTrue)
		})
		Convey("DEBUG=\"tRuE\" enables debugging", func() {
			os.Setenv("DEBUG", "tRuE")
			os.Args = []string{"spruce"}
			DebugOn = false
			main()
			So(DebugOn, ShouldBeTrue)
		})
		Convey("DEBUG=1 enables debugging", func() {
			os.Setenv("DEBUG", "1")
			os.Args = []string{"spruce"}
			DebugOn = false
			main()
			So(DebugOn, ShouldBeTrue)
		})
		Convey("DEBUG=randomval enables debugging", func() {
			os.Setenv("DEBUG", "randomval")
			os.Args = []string{"spruce"}
			DebugOn = false
			main()
			So(DebugOn, ShouldBeTrue)
		})
		Convey("DEBUG=\"fAlSe\" disables debugging", func() {
			os.Setenv("DEBUG", "fAlSe")
			os.Args = []string{"spruce"}
			DebugOn = false
			main()
			So(DebugOn, ShouldBeFalse)
		})
		Convey("DEBUG=0 disables debugging", func() {
			os.Setenv("DEBUG", "0")
			os.Args = []string{"spruce"}
			DebugOn = false
			main()
			So(DebugOn, ShouldBeFalse)
		})
		Convey("DEBUG=\"\" disables debugging", func() {
			os.Setenv("DEBUG", "")
			os.Args = []string{"spruce"}
			DebugOn = false
			main()
			So(DebugOn, ShouldBeFalse)
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

func TestExamples(t *testing.T) {
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

	YAML := func(path string) string {
		s, err := ioutil.ReadFile(path)
		So(err, ShouldBeNil)

		y, err := simpleyaml.NewYaml([]byte(s))
		So(err, ShouldBeNil)

		data, err := y.Map()
		So(err, ShouldBeNil)

		out, err := yaml.Marshal(data)
		So(err, ShouldBeNil)

		return string(out) + "\n"
	}

	Convey("Examples from README.md", t, func() {
		example := func(args ...string) {
			expect := args[len(args)-1]
			args = args[:len(args)-1]

			os.Args = []string{"spruce", "merge"}
			os.Args = append(os.Args, args...)
			stdout, stderr = "", ""
			main()

			So(stderr, ShouldEqual, "")
			So(stdout, ShouldEqual, YAML(expect))
		}

		Convey("Basic Example", func() {
			example(
				"../../examples/basic/main.yml",
				"../../examples/basic/merge.yml",

				"../../examples/basic/output.yml",
			)
		})

		Convey("Map Replacements", func() {
			example(
				"../../examples/map-replacement/original.yml",
				"../../examples/map-replacement/delete.yml",
				"../../examples/map-replacement/insert.yml",

				"../../examples/map-replacement/output.yml",
			)
		})

		Convey("Key Removal", func() {
			example(
				"--prune", "deleteme",
				"../../examples/key-removal/original.yml",
				"../../examples/key-removal/things.yml",

				"../../examples/key-removal/output.yml",
			)

			example(
				"../../examples/pruning/base.yml",
				"../../examples/pruning/jobs.yml",
				"../../examples/pruning/networks.yml",

				"../../examples/pruning/output.yml",
			)
		})

		Convey("Lists of Maps", func() {
			example(
				"../../examples/list-of-maps/original.yml",
				"../../examples/list-of-maps/new.yml",

				"../../examples/list-of-maps/output.yml",
			)
		})

		Convey("Static IPs", func() {
			example(
				"../../examples/static-ips/jobs.yml",
				"../../examples/static-ips/properties.yml",
				"../../examples/static-ips/networks.yml",

				"../../examples/static-ips/output.yml",
			)
		})

		Convey("Static IPs with availability zones", func() {
			example(
				"../../examples/availability-zones/jobs.yml",
				"../../examples/availability-zones/properties.yml",
				"../../examples/availability-zones/networks.yml",

				"../../examples/availability-zones/output.yml",
			)
		})

		Convey("Injecting Subtrees", func() {
			example(
				"--prune", "meta",
				"../../examples/inject/all-in-one.yml",

				"../../examples/inject/output.yml",
			)

			example(
				"--prune", "meta",
				"../../examples/inject/templates.yml",
				"../../examples/inject/green.yml",

				"../../examples/inject/output.yml",
			)
		})

		Convey("Pruning", func() {
			example(
				"../../examples/pruning/base.yml",
				"../../examples/pruning/jobs.yml",
				"../../examples/pruning/networks.yml",

				"../../examples/pruning/output.yml",
			)
		})

		Convey("Inserting", func() {
			example(
				"../../examples/inserting/main.yml",
				"../../examples/inserting/addon.yml",

				"../../examples/inserting/result.yml",
			)
		})

		Convey("Calc", func() {
			example(
				"--prune", "meta",
				"../../examples/calc/meta.yml",
				"../../examples/calc/jobs.yml",

				"../../examples/calc/result.yml",
			)
		})
	})
}
