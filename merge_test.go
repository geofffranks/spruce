package main

import "testing"
import . "github.com/smartystreets/goconvey/convey"

func TestShouldMergeArray(t *testing.T) {
	Convey("We should merge arrays", t, func() {
		Convey("If the element is a string with the right token", func() {
			So(shouldMergeArray("(( merge ))"), ShouldBeTrue)
		})
		Convey("But not if the element is a string with the wrong token", func() {
			So(shouldMergeArray("not a magic token"), ShouldBeFalse)
		})
		Convey("But not if the element is not a string", func() {
			So(shouldMergeArray(42), ShouldBeFalse)
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
		orig_map := map[interface{}]interface{}{"k1": "v1", "k2": "v2"}
		new_map := map[interface{}]interface{}{"k3": "v3", "k2": "v2.new"}
		expect_map := map[interface{}]interface{}{"k2": "v2.new", "k3": "v3", "k1": "v1"}

		mergeMap(orig_map, new_map, "node-path")
		So(orig_map, ShouldResemble, expect_map)
	})
}

func TestMergeArray(t *testing.T) {
	Convey("mergeArray", t, func() {
		Convey("with initial element '(( merge ))' appends new data", func() {
			orig := []interface{}{"first", "second"}
			array := []interface{}{"zeroth", "(( merge ))"}
			expect := []interface{}{"zeroth", "first", "second"}

			So(mergeObj(orig, array, "node-path"), ShouldResemble, expect)
		})
		Convey("with final element '(( merge ))' prepends new data", func() {
			orig := []interface{}{"first", "second"}
			array := []interface{}{"(( merge ))", "third"}
			expect := []interface{}{"first", "second", "third"}

			So(mergeObj(orig, array, "node-path"), ShouldResemble, expect)
		})
		Convey("with map elements replaces entire array", func() {
			orig_mapslice := map[string]string{"k1": "v1", "k2": "v2"}
			new_mapslice := map[string]string{"k3": "v3", "k2": "v2.new"}
			expect_mapslice := map[string]string{"k2": "v2.new", "k3": "v3"}
			orig := []interface{}{orig_mapslice}
			array := []interface{}{new_mapslice}
			expect := []interface{}{expect_mapslice}

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
