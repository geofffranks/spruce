package spruce

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/starkandwayne/goutils/tree"
)

func TestSort(t *testing.T) {
	Convey("that the actual Run function of the sort operator is dead code", t, func() {
		op := &SortOperator{}
		ev := &Evaluator{Here: &tree.Cursor{Nodes: []string{"foobar"}}}

		resp, err := op.Run(ev, []*Expr{})
		So(resp, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldResemble, "orphaned (( sort )) operator at $.foobar, no list exists at that path")
	})

	Convey("that sorting an empty list returns an empty list", t, func() {
		list := []interface{}{}
		err := sortList("some.path", list, "")
		So(err, ShouldBeNil)
		So(list, ShouldResemble, []interface{}{})
	})

	Convey("that sorting of integers works", t, func() {
		list := []interface{}{2, 1}
		err := sortList("some.path", list, "")
		So(err, ShouldBeNil)
		So(list, ShouldResemble, []interface{}{1, 2})
	})

	Convey("that sorting of floats works", t, func() {
		list := []interface{}{2.0, 1.0}
		err := sortList("some.path", list, "")
		So(err, ShouldBeNil)
		So(list, ShouldResemble, []interface{}{1.0, 2.0})
	})

	Convey("that sorting of strings works", t, func() {
		list := []interface{}{"spruce", "spiff"}
		err := sortList("some.path", list, "")
		So(err, ShouldBeNil)
		So(list, ShouldResemble, []interface{}{"spiff", "spruce"})
	})

	Convey("that sorting of named-entry lists works", t, func() {
		list := []interface{}{
			map[interface{}]interface{}{"name": "B"},
			map[interface{}]interface{}{"name": "A"},
		}
		err := sortList("some.path", list, "")
		So(err, ShouldBeNil)
		So(list, ShouldResemble, []interface{}{
			map[interface{}]interface{}{"name": "A"},
			map[interface{}]interface{}{"name": "B"},
		})
	})

	Convey("that sorting of lists of lists fails", t, func() {
		list := []interface{}{
			[]interface{}{"B", "A"},
			[]interface{}{"A", "B"},
		}

		err := sortList("some.path", list, "")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldResemble, "$.some.path is a list with list entries (not a list with maps, strings or numbers)")
	})

	Convey("that sorting of a list of inhomogeneous types fails", t, func() {
		list := []interface{}{42, 42.0, "42"}
		err := sortList("some.path", list, "")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldResemble, "$.some.path is a list with different types (not a list with homogeneous entry types)")
	})

	Convey("that sorting of a list with nil values fails (by definition considered to be inhomogeneous types)", t, func() {
		list := []interface{}{"A", "B", "C", nil}
		err := sortList("some.path", list, "")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldResemble, "$.some.path is a list with different types (not a list with homogeneous entry types)")
	})

	Convey("that sorting of a named-entry list with inconsistent identifier fails", t, func() {
		list := []interface{}{
			map[interface{}]interface{}{"foo": "one"},
			map[interface{}]interface{}{"key": "two"},
		}
		err := sortList("some.path", list, "foo")
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldResemble, "$.some.path is a list with map entries, where some do not contain foo (not a list with map entries each containing foo)")
	})
}
