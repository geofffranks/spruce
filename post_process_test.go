package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPostProcessMap(t *testing.T) {
	Convey("postProcessMap()", t, func() {
		Convey("returns errors when postProcessObj() fails", func() {
			data := map[interface{}]interface{}{
				"reference": "(( grab value.to.reference ))",
			}
			root := map[interface{}]interface{}{
				"value": map[interface{}]interface{}{
					"to": map[interface{}]interface{}{},
				},
			}

			err := postProcessMap(data, root, "nodepath")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "nodepath.reference: Unable to resolve `value.to.reference`")
		})
		Convey("Under normal circumstances", func() {
			data := map[interface{}]interface{}{
				"leave":     "me alone",
				"reference": "(( grab value.to.reference ))",
				"nilval":    "(( grab value.nilval ))",
			}
			root := map[interface{}]interface{}{
				"value": map[interface{}]interface{}{
					"to": map[interface{}]interface{}{
						"reference": "I referenced the value!",
					},
					"nilval": nil,
				},
			}
			err := postProcessMap(data, root, "nodepath")
			Convey("returns nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Updates the node(s) with references (and nothing else)", func() {
				expect := map[interface{}]interface{}{
					"leave":     "me alone",
					"reference": "I referenced the value!",
					"nilval":    nil,
				}
				So(data, ShouldResemble, expect)
			})
		})
	})
}
func TestPostProcessArray(t *testing.T) {
	Convey("postProcessArray()", t, func() {
		Convey("returns errors when postProcessObj() fails", func() {
			data := []interface{}{
				"(( grab value.to.reference ))",
			}
			root := map[interface{}]interface{}{
				"value": map[interface{}]interface{}{
					"to": map[interface{}]interface{}{},
				},
			}

			err := postProcessArray(data, root, "nodepath")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "nodepath.[0]: Unable to resolve `value.to.reference`")
		})
		Convey("Under normal circumstances", func() {
			data := []interface{}{
				"leave me alone",
				"(( grab value.to.reference ))",
				"(( grab value.nil ))",
			}
			root := map[interface{}]interface{}{
				"value": map[interface{}]interface{}{
					"to": map[interface{}]interface{}{
						"reference": "I referenced the value!",
					},
					"nil": nil,
				},
			}

			expect := []interface{}{
				"leave me alone",
				"I referenced the value!",
				nil,
			}
			err := postProcessArray(data, root, "nodepath")
			Convey("returns nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Updates the node(s) with references (and nothing else)", func() {
				So(data, ShouldResemble, expect)
			})
		})
	})
}

func TestPostProcessObj(t *testing.T) {
	Convey("postProcessObj()", t, func() {
		Convey("When passed a map", func() {
			Convey("Post-processes as a map under normal conditions", func() {
				data := map[interface{}]interface{}{
					"leave":     "me alone",
					"reference": "(( grab value.to.reference ))",
					"nilval":    "(( grab value.nil ))",
				}
				root := map[interface{}]interface{}{
					"value": map[interface{}]interface{}{
						"to": map[interface{}]interface{}{
							"reference": "I referenced the value!",
						},
						"nil": nil,
					},
				}

				expect := map[interface{}]interface{}{
					"leave":     "me alone",
					"reference": "I referenced the value!",
					"nilval":    nil,
				}
				got, err := postProcessObj(data, root, "nodepath")
				Convey("returns no error", func() {
					So(err, ShouldBeNil)
				})
				Convey("Updates the node(s) with references (and nothing else)", func() {
					So(got, ShouldResemble, expect)
				})
			})
			Convey("Returns an error when postProcessMap() throws an error", func() {
				data := map[interface{}]interface{}{
					"reference": "(( grab value.to.reference ))",
				}
				root := map[interface{}]interface{}{
					"value": map[interface{}]interface{}{
						"to": map[interface{}]interface{}{},
					},
				}

				got, err := postProcessObj(data, root, "nodepath")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldStartWith, "nodepath.reference: Unable to resolve `value.to.reference`")
				So(got, ShouldBeNil)
			})
		})
		Convey("When passed an array", func() {
			Convey("Post-processes as an array under normal conditions", func() {
				data := []interface{}{
					"leave me alone",
					"(( grab value.to.reference ))",
					"(( grab value.nil ))",
				}
				root := map[interface{}]interface{}{
					"value": map[interface{}]interface{}{
						"to": map[interface{}]interface{}{
							"reference": "I referenced the value!",
						},
						"nil": nil,
					},
				}

				expect := []interface{}{
					"leave me alone",
					"I referenced the value!",
					nil,
				}
				got, err := postProcessObj(data, root, "nodepath")
				Convey("returns nil", func() {
					So(err, ShouldBeNil)
				})
				Convey("Updates the node(s) with references (and nothing else)", func() {
					So(got, ShouldResemble, expect)
				})
			})
			Convey("Returns an error when postProcessArray() throws an error", func() {
				data := []interface{}{
					"(( grab value.to.reference ))",
				}
				root := map[interface{}]interface{}{
					"value": map[interface{}]interface{}{
						"to": map[interface{}]interface{}{},
					},
				}

				val, err := postProcessObj(data, root, "nodepath")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldStartWith, "nodepath.[0]: Unable to resolve `value.to.reference`")
				So(val, ShouldBeNil)
			})
		})
		Convey("When passed a string", func() {
			root := map[interface{}]interface{}{
				"value": map[interface{}]interface{}{
					"to": map[interface{}]interface{}{
						"reference": "referenced value",
					},
				},
			}
			Convey("That matches the resolve token", func() {
				Convey("returns resolved data if no issues were encountered", func() {
					got, err := postProcessObj("(( grab value.to.reference ))", root, "nodepath")
					So(got, ShouldEqual, "referenced value")
					So(err, ShouldBeNil)
				})
				Convey("Returns nil, and an error if issues were encountered resolving", func() {
					got, err := postProcessObj("(( grab stuff ))", root, "nodepath")
					So(got, ShouldBeNil)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldStartWith, "nodepath: Unable to resolve `stuff`")
				})
			})
			Convey("That doesn't match the resolve token", func() {
				Convey("Returns nil, and an appropriate error", func() {
					got, err := postProcessObj("13", root, "nodepath")
					So(got, ShouldBeNil)
					So(err.Error(), ShouldEqual, "nodepath: does not need to be resolved")
				})
			})
		})
		Convey("When passed anything else", func() {
			root := map[interface{}]interface{}{
				"value": map[interface{}]interface{}{
					"to": map[interface{}]interface{}{
						"reference": "referenced value",
					},
				},
			}
			Convey("Returns nil, and an appropriate error", func() {
				got, err := postProcessObj(13, root, "nodepath")
				So(got, ShouldBeNil)
				So(err.Error(), ShouldEqual, "nodepath: does not need to be resolved")
			})
		})
	})
}

func TestShouldResolveString(t *testing.T) {
	Convey("shouldResolveString()", t, func() {
		Convey("returns resolve text + true on match", func() {
			str, should := shouldResolveString("(( grab my.property ))")
			So(str, ShouldEqual, "my.property")
			So(should, ShouldBeTrue)
		})
		Convey("front and back whitespace optional", func() {
			str, should := shouldResolveString("((grab my.property))")
			So(str, ShouldEqual, "my.property")
			So(should, ShouldBeTrue)
		})
		Convey("front and back whitespace can be huge", func() {
			str, should := shouldResolveString("((    grab   my.property		))")
			So(str, ShouldEqual, "my.property")
			So(should, ShouldBeTrue)
		})
		Convey("returns empty text and false on no match", func() {
			str, should := shouldResolveString("my.property")
			So(str, ShouldEqual, "")
			So(should, ShouldBeFalse)
		})
		Convey("Is anchored such that quoting (( .* )) returns no match", func() {
			str, should := shouldResolveString("\"(( grab my.property ))\"")
			So(str, ShouldEqual, "")
			So(should, ShouldBeFalse)
		})
	})
}

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
