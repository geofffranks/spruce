package main

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type MockPostProcessor struct {
	action string
	value  interface{}
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

func TestWalkTree(t *testing.T) {
	Convey("Visit()", t, func() {
		m := &Merger{}
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
		Convey("Sets node to dollar-sign", func() {
			Convey("When node is empty string", func() {
				tree["error"] = "(( mock ))"
				m.Visit(tree, MockPostProcessor{action: "error"})
				err := m.Error()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "$.error: fake error")
			})
			Convey("but not when node is specified", func() {
				tree["error"] = "(( mock ))"
				m.Visit(tree, MockPostProcessor{action: "error"})
				err := m.Error()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "error")
				So(err.Error(), ShouldContainSubstring, ": fake error")
			})
		})
		Convey("Bails out if recursion gets too high", func() {
			tree["recurse"] = tree
			m.Visit(tree, MockPostProcessor{action: "ignore"})
			err := m.Error()
			delete(tree, "recurse")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "hit max recursion depth. You seem to have a self-referencing dataset")
		})
		Convey("Replaces values in maps if postprocessor told it to", func() {
			Convey("Regular values are just assigned", func() {
				tree["replaceme"] = "(( mock ))"
				m.Visit(tree, MockPostProcessor{action: "replace", value: "1234"})
				err := m.Error()
				So(err, ShouldBeNil)
				So(tree["replaceme"], ShouldEqual, "1234")
				delete(tree, "replaceme")
			})
			Convey("Maps are deep-copied", func() {
				newMap := map[interface{}]interface{}{
					"newKey": "newVal",
				}
				tree["replaceme"] = "(( mock ))"
				m.Visit(tree, MockPostProcessor{action: "replace", value: newMap})
				err := m.Error()
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
				m.Visit(tree, MockPostProcessor{action: "replace", value: "1234"})
				err := m.Error()
				So(err, ShouldBeNil)
				So(tree["replaceme"].([]interface{})[2], ShouldEqual, "1234")
				delete(tree, "replaceme")
			})
			Convey("Maps are deep-copied", func() {
				tree["replaceme"] = array
				newMap := map[interface{}]interface{}{
					"newKey": "newVal",
				}
				m.Visit(tree, MockPostProcessor{action: "replace", value: newMap})
				err := m.Error()
				So(err, ShouldBeNil)
				So(tree["replaceme"].([]interface{})[2], ShouldResemble, newMap)
				So(tree["replaceme"].([]interface{})[2], ShouldNotEqual, newMap)
				delete(tree, "replaceme")
			})
		})
		Convey("Does nothing for values postprocessor ignores", func() {
			m.Visit(tree, MockPostProcessor{action: "replace", value: 1})
			err := m.Error()
			So(tree["color"], ShouldEqual, "blue")
			So(err, ShouldBeNil)
		})
		Convey("Returns errors if the postprocessor has a problem", func() {
			Convey("when recursing over maps", func() {
				tree["error"] = "(( mock ))"
				m.Visit(tree, MockPostProcessor{action: "error"})
				err := m.Error()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "$.error: fake error")
				delete(tree, "error")
			})
			Convey("when recursing over arrays", func() {
				tree["error"] = []interface{}{
					1,
					2,
					"(( mock ))",
					3,
				}
				m.Visit(tree, MockPostProcessor{action: "error"})
				err := m.Error()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "$.error.[2]: fake error")
				delete(tree, "error")
			})
		})
	})
}
