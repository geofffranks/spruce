package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestShouldAppendToArray(t *testing.T) {
	Convey("We should append to arrays", t, func() {
		Convey("If the element is a string with the right append token", func() {
			So(shouldAppendToArray([]interface{}{"(( append ))", "stuff"}), ShouldBeTrue)
		})
		Convey("But not if the element is a string with the wrong token", func() {
			So(shouldAppendToArray([]interface{}{"not a magic token"}), ShouldBeFalse)
		})
		Convey("But not if the element is not a string", func() {
			So(shouldAppendToArray([]interface{}{42}), ShouldBeFalse)
		})
		Convey("But not if the slice has no elements", func() {
			So(shouldInlineMergeArray([]interface{}{}), ShouldBeFalse)
		})
	})
}
func TestShouldPrependArray(t *testing.T) {
	Convey("We should prepend to arrays", t, func() {
		Convey("If the element is a string with the right prepend token", func() {
			So(shouldPrependToArray([]interface{}{"(( prepend ))", "stuff"}), ShouldBeTrue)
		})
		Convey("But not if the element is a string with the wrong token", func() {
			So(shouldPrependToArray([]interface{}{"not a magic token"}), ShouldBeFalse)
		})
		Convey("But not if the element is not a string", func() {
			So(shouldPrependToArray([]interface{}{42}), ShouldBeFalse)
		})
		Convey("But not if the slice has no elements", func() {
			So(shouldInlineMergeArray([]interface{}{}), ShouldBeFalse)
		})
	})
}
func TestShouldInlineMergeArray(t *testing.T) {
	Convey("We should inline merge arrays", t, func() {
		Convey("If the element is a string with the right inline-merge token", func() {
			So(shouldInlineMergeArray([]interface{}{"(( inline ))", "stuff"}), ShouldBeTrue)
		})
		Convey("But not if the element is a string with the wrong token", func() {
			So(shouldInlineMergeArray([]interface{}{"not a magic token"}), ShouldBeFalse)
		})
		Convey("But not if the element is not a string", func() {
			So(shouldInlineMergeArray([]interface{}{42}), ShouldBeFalse)
		})
		Convey("But not if the slice has no elements", func() {
			So(shouldInlineMergeArray([]interface{}{}), ShouldBeFalse)
		})
	})
}

func TestMergeObj(t *testing.T) {
	Convey("Passing a map to mergeObj merges as a map", t, func() {

	})
	Convey("Passing a slice to mergeObj", t, func() {
		Convey("without magical merge token replaces entire array", func() {
			orig := []interface{}{"first", "second"}
			array := []interface{}{"my", "new", "array"}
			expect := []interface{}{"my", "new", "array"}

			So(mergeObj(orig, array, "node-path"), ShouldResemble, expect)
		})
	})
	Convey("mergeObj merges in place", t, func() {
		Convey("When passed a string", func() {
			orig := 42
			val := "asdf"
			So(mergeObj(orig, val, "node-path"), ShouldEqual, "asdf")
		})
		Convey("When passed an int", func() {
			orig := "fdsa"
			val := 10
			So(mergeObj(orig, val, "node-path"), ShouldEqual, 10)
		})
		Convey("When passed an float64", func() {
			orig := "fdsa"
			val := 10.4
			So(mergeObj(orig, val, "node-path"), ShouldEqual, 10.4)
		})
		Convey("When passed nil", func() {
			orig := "fdsa"
			val := interface{}(nil)
			So(mergeObj(orig, val, "node-path"), ShouldBeNil)
		})
		Convey("When passed a map, but original item is a scalar", func() {
			orig := "value"
			val := map[interface{}]interface{}{"key": "value"}
			expect := map[interface{}]interface{}{"key": "value"}
			So(mergeObj(orig, val, "node-path"), ShouldResemble, expect)
		})
		Convey("When passed a map, but original item is nil", func() {
			val := map[interface{}]interface{}{"key": "value"}
			expect := map[interface{}]interface{}{"key": "value"}
			So(mergeObj(nil, val, "node-path"), ShouldResemble, expect)
		})
		Convey("When passed a slice, but original item is a scalar", func() {
			orig := "value"
			val := []interface{}{"one", "two"}
			expect := []interface{}{"one", "two"}
			So(mergeObj(orig, val, "node-path"), ShouldResemble, expect)
		})
		Convey("When passed a slice, but original item is nil", func() {
			val := []interface{}{"one", "two"}
			expect := []interface{}{"one", "two"}
			So(mergeObj(nil, val, "node-path"), ShouldResemble, expect)
		})
	})
}

func TestMergeMap(t *testing.T) {
	Convey("with map elements updates original map", t, func() {
		origMap := map[interface{}]interface{}{"k1": "v1", "k2": "v2"}
		newMap := map[interface{}]interface{}{"k3": "v3", "k2": "v2.new"}
		expectMap := map[interface{}]interface{}{"k2": "v2.new", "k3": "v3", "k1": "v1"}

		mergeMap(origMap, newMap, "node-path")
		So(origMap, ShouldResemble, expectMap)
	})
}

func TestMergeArray(t *testing.T) {
	Convey("mergeArray", t, func() {
		Convey("with initial element '(( prepend ))' prepends new data", func() {
			orig := []interface{}{"first", "second"}
			array := []interface{}{"(( prepend ))", "zeroth"}
			expect := []interface{}{"zeroth", "first", "second"}

			So(mergeArray(orig, array, "node-path"), ShouldResemble, expect)
		})
		Convey("with initial element '(( append ))' appends new data", func() {
			orig := []interface{}{"first", "second"}
			array := []interface{}{"(( append ))", "third"}
			expect := []interface{}{"first", "second", "third"}

			So(mergeArray(orig, array, "node-path"), ShouldResemble, expect)
		})
		Convey("with initial element '(( inline ))'", func() {
			Convey("and len(orig) == len(new)", func() {
				orig := []interface{}{
					"orig_first",
					[]interface{}{"subfirst", "subsecond"},
					map[interface{}]interface{}{
						"name": "original",
						"val":  "original",
					},
					"orig_last",
				}
				array := []interface{}{
					"(( inline ))",
					"overwritten_first",
					[]interface{}{"(( append ))", "subthird"},
					map[interface{}]interface{}{
						"val": "overwritten",
					},
					"overwritten_last",
				}
				expect := []interface{}{
					"overwritten_first",
					[]interface{}{"subfirst", "subsecond", "subthird"},
					map[interface{}]interface{}{
						"name": "original",
						"val":  "overwritten",
					},
					"overwritten_last",
				}
				So(mergeArray(orig, array, "node-path"), ShouldResemble, expect)
			})
			Convey("and len(orig) > len(new)", func() {
				orig := []interface{}{
					"orig_first",
					[]interface{}{"subfirst", "subsecond"},
					map[interface{}]interface{}{
						"name": "original",
						"val":  "original",
					},
					"orig_last",
				}
				array := []interface{}{
					"(( inline ))",
					"overwritten_first",
					[]interface{}{"(( append ))", "subthird"},
					map[interface{}]interface{}{
						"val": "overwritten",
					},
				}
				expect := []interface{}{
					"overwritten_first",
					[]interface{}{"subfirst", "subsecond", "subthird"},
					map[interface{}]interface{}{
						"name": "original",
						"val":  "overwritten",
					},
					"orig_last",
				}
				So(mergeArray(orig, array, "node-path"), ShouldResemble, expect)
			})
			Convey("and len(orig < len(new)", func() {
				orig := []interface{}{
					"orig_first",
					[]interface{}{"subfirst", "subsecond"},
				}
				array := []interface{}{
					"(( inline ))",
					"overwritten_first",
					[]interface{}{"(( append ))", "subthird"},
					map[interface{}]interface{}{
						"val": "overwritten",
					},
					"overwritten_last",
				}
				expect := []interface{}{
					"overwritten_first",
					[]interface{}{"subfirst", "subsecond", "subthird"},
					map[interface{}]interface{}{
						"val": "overwritten",
					},
					"overwritten_last",
				}
				So(mergeArray(orig, array, "node-path"), ShouldResemble, expect)
			})
			Convey("and empty orig slice", func() {
			})
			Convey("and empty new slice", func() {
			})
		})
		Convey("with map elements replaces entire array", func() {
			origMapSlice := map[string]string{"k1": "v1", "k2": "v2"}
			newMapSlice := map[string]string{"k3": "v3", "k2": "v2.new"}
			expectMapSlice := map[string]string{"k2": "v2.new", "k3": "v3"}
			orig := []interface{}{origMapSlice}
			array := []interface{}{newMapSlice}
			expect := []interface{}{expectMapSlice}

			So(mergeObj(orig, array, "node-path"), ShouldResemble, expect)
		})
		Convey("without magical merge token replaces entire array", func() {
			orig := []interface{}{"first", "second"}
			array := []interface{}{"my", "new", "array"}
			expect := []interface{}{"my", "new", "array"}

			So(mergeObj(orig, array, "node-path"), ShouldResemble, expect)
		})
	})
}
