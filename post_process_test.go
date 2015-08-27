package main

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

/*
func TestDeepCopy(t *testing.T) {
	Convey("deepCopy()", t, func() {
		Convey("Makes a deep clone of a map", func() {
			data := map[string]string{
				"key": "value",
			}
			got := make(map[string]string)
			deepCopy(got, data)
			So(got, ShouldResemble, data)
			So(got, ShouldNotEqual, data)
		})
	})
}
*/
func TestDeReferencerAction(t *testing.T) {
	Convey("dereferencer.Action() returns correct string", t, func() {
		deref := DeReferencer{root: map[interface{}]interface{}{}}
		So(deref.Action(), ShouldEqual, "dereference")
	})
}

func TestDeReferencerPostProcess(t *testing.T) {
	Convey("dereferencer.PostProces()", t, func() {
		deref := DeReferencer{root: map[interface{}]interface{}{
			"value": map[interface{}]interface{}{
				"to": map[interface{}]interface{}{
					"find": "dereferenced value",
				},
			},
		}}
		Convey("returns nil, \"ignore\", nil", func() {
			Convey("when given anything other than a string", func() {
				val, action, err := deref.PostProcess(12345, "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
			Convey("when given a '(( prune ))' string", func() {
				val, action, err := deref.PostProcess("(( prune ))", "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
			Convey("when given a non-'(( grab .* ))' string", func() {
				val, action, err := deref.PostProcess("regular old string", "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
			Convey("when given a quoted-'(( grab .* ))' string", func() {
				val, action, err := deref.PostProcess("\"(( grab value.to.find ))\"", "nodepath")
				So(val, ShouldBeNil)
				So(err, ShouldBeNil)
				So(action, ShouldEqual, "ignore")
			})
		})
		Convey("Returns an error if resolveNode() had an error resolving", func() {
			val, action, err := deref.PostProcess("(( grab value.to.retrieve ))", "nodepath")
			So(val, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldStartWith, "nodepath: Unable to resolve `value.to.retrieve`:")
			So(action, ShouldEqual, "error")
		})
		Convey("Returns value, \"replace\", nil on successful dereference", func() {
			val, action, err := deref.PostProcess("(( grab value.to.find ))", "nodepath")
			So(val, ShouldEqual, "dereferenced value")
			So(err, ShouldBeNil)
			So(action, ShouldEqual, "replace")
		})
	})
}

type MockPostProcessor struct {
	action string
	value  interface{}
}

func (p MockPostProcessor) Action() string {
	return p.action
}
func (p MockPostProcessor) PostProcess(i interface{}, node string) (interface{}, string, error) {
	if _, ok := i.(string); ok && i.(string) == "(( mock ))" {
		if p.action == "error" {
			return nil, "error", fmt.Errorf("%s: fake error", node)
		}
		if p.action == "replace" {
			return p.value, "replace", nil
		}
	}
	return nil, "ignore", nil
}

func TestWalkTree(t *testing.T) {
	Convey("walkTree()", t, func() {
		tree := map[interface{}]interface{}{
			"color": "blue",
			"array": []interface{}{
				1,
				2,
				map[interface{}]interface{}{
					"shape": "rectangle",
				},
			},
			"map": map[interface{}]interface{}{
				"s-car": "go",
			},
		}
		Convey("Resets CURRENT_DEPTH", func() {
			Convey("When node is ''", func() {
				CURRENT_DEPTH = 10
				walkTree(tree, MockPostProcessor{action: "ignore"}, "")
				So(CURRENT_DEPTH, ShouldEqual, 0)
			})
			Convey("but not when node is specified", func() {
				CURRENT_DEPTH = 10
				walkTree(tree, MockPostProcessor{action: "ignore"}, "node")
				So(CURRENT_DEPTH, ShouldEqual, 10)
			})
		})
		Convey("Sets node to dollar-sign", func() {
			Convey("When node is empty string", func() {
				tree["error"] = "(( mock ))"
				err := walkTree(tree, MockPostProcessor{action: "error"}, "")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "$.error: fake error")
			})
			Convey("but not when node is specified", func() {
				tree["error"] = "(( mock ))"
				err := walkTree(tree, MockPostProcessor{action: "error"}, "node")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldStartWith, "node.error")
				So(err.Error(), ShouldEndWith, ": fake error")
			})
		})
		Convey("Bails out if recursion gets too high", func() {
			tree["recurse"] = tree
			err := walkTree(tree, MockPostProcessor{action: "ignore"}, "")
			delete(tree, "recurse")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEndWith, "hit max recursion depth. You seem to have a self-referencing dataset.")
		})
		Convey("Replaces values in maps if postprocessor told it to", func() {
			Convey("Regular values are just assigned", func() {
				tree["replaceme"] = "(( mock ))"
				err := walkTree(tree, MockPostProcessor{action: "replace", value: "1234"}, "")
				So(err, ShouldBeNil)
				So(tree["replaceme"], ShouldEqual, "1234")
				delete(tree, "replaceme")
			})
			Convey("Maps are deep-copied", func() {
				newMap := map[interface{}]interface{}{
					"newKey": "newVal",
				}
				tree["replaceme"] = "(( mock ))"
				err := walkTree(tree, MockPostProcessor{action: "replace", value: newMap}, "")
				So(err, ShouldBeNil)
				So(tree["replaceme"], ShouldResemble, newMap)
				So(tree["replaceme"], ShouldNotEqual, newMap)
				delete(tree, "replaceme")
			})
		})
		Convey("Replaces values in arrays if postprocessor told it to", func() {
			array := []interface{}{
				1,
				2,
				"(( mock ))",
				3,
			}
			Convey("Regular values are just assigned", func() {
				tree["replaceme"] = array
				err := walkTree(tree, MockPostProcessor{action: "replace", value: "1234"}, "")
				So(err, ShouldBeNil)
				So(tree["replaceme"].([]interface{})[2], ShouldEqual, "1234")
				delete(tree, "replaceme")
			})
			Convey("Maps are deep-copied", func() {
				tree["replaceme"] = array
				newMap := map[interface{}]interface{}{
					"newKey": "newVal",
				}
				err := walkTree(tree, MockPostProcessor{action: "replace", value: newMap}, "")
				So(err, ShouldBeNil)
				So(tree["replaceme"].([]interface{})[2], ShouldResemble, newMap)
				So(tree["replaceme"].([]interface{})[2], ShouldNotEqual, newMap)
				delete(tree, "replaceme")
			})
		})
		Convey("Does nothing for values postprocessor ignores", func() {
			err := walkTree(tree, MockPostProcessor{action: "replace", value: 1}, "")
			So(tree["color"], ShouldEqual, "blue")
			So(err, ShouldBeNil)
		})
		Convey("Returns errors if the postprocessor has a problem", func() {
			Convey("when recursing over maps", func() {
				tree["error"] = "(( mock ))"
				err := walkTree(tree, MockPostProcessor{action: "error"}, "")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "$.error: fake error")
				delete(tree, "error")
			})
			Convey("when recursing over arrays", func() {
				tree["error"] = []interface{}{
					1,
					2,
					"(( mock ))",
					3,
				}
				err := walkTree(tree, MockPostProcessor{action: "error"}, "")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "$.error.[2]: fake error")
				delete(tree, "error")
			})
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
