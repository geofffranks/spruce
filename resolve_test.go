package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestResolveNode(t *testing.T) {
	Convey("resolveNode()", t, func() {
		Convey("Finds a value for a given map path", func() {
			m := map[interface{}]interface{}{
				"value": map[interface{}]interface{}{
					"to": map[interface{}]interface{}{
						"resolve": "I referenced the property!",
					},
				},
				"other": "stuff",
			}
			expect := "I referenced the property!"
			val, err := resolveNode("value.to.resolve", m)
			So(err, ShouldBeNil)
			So(val, ShouldResemble, expect)
		})
		Convey("Finds a value for an index-based array path", func() {
			m := map[interface{}]interface{}{
				"array": []interface{}{
					"first",
					"second",
					"third",
				},
			}

			expect := "third"
			val, err := resolveNode("array.[2]", m)
			So(err, ShouldBeNil)
			So(val, ShouldResemble, expect)
		})
		Convey("Finds a value for a name-based array path", func() {
			m := map[interface{}]interface{}{
				"jobs": []interface{}{
					map[interface{}]interface{}{
						"name": "job1",
						"processes": []interface{}{
							"nginx",
							"crond",
							"monit",
						},
					},
					map[interface{}]interface{}{
						"name": "job2",
						"processes": []interface{}{
							"java",
							"oomkiller",
						},
					},
				},
			}
			expect := []interface{}{
				"java",
				"oomkiller",
			}
			val, err := resolveNode("jobs.job2.processes", m)
			So(err, ShouldBeNil)
			So(val, ShouldResemble, expect)
		})
		Convey("Returns an error for invalid array indices", func() {
			m := map[interface{}]interface{}{
				"jobs": []interface{}{
					"first",
					"second",
				},
			}
			val, err := resolveNode("jobs.[2]", m)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.[2]` array's highest index is 1")
			So(val, ShouldBeNil)
		})
		Convey("Returns an error for referencing non-map/array nodes", func() {
			m := map[interface{}]interface{}{
				"jobs": []interface{}{
					map[interface{}]interface{}{
						"name":     "job1",
						"property": "scalar",
					},
				},
			}
			val, err := resolveNode("jobs.job1.property.myval", m)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.job1.property` has no sub-objects")
			So(val, ShouldBeNil)
		})
		Convey("Returns an error for referencing a key that does not exist in a map", func() {
			m := map[interface{}]interface{}{
				"existing": "value",
			}
			val, err := resolveNode("non.existant.value", m)
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "non` could not be found in the YAML datastructure")
		})
		Convey("Returns an error for referencing an element object that does not exist in an array", func() {
			m := map[interface{}]interface{}{
				"jobs": []interface{}{
					map[interface{}]interface{}{
						"name":     "job1",
						"property": "scalar",
					},
				},
			}
			val, err := resolveNode("jobs.job2.value", m)
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "jobs.job2` could not be found in the YAML datastructure")
		})
	})
}
