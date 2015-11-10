package main

import (
	"github.com/geofffranks/simpleyaml" // FIXME: switch back to smallfish/simpleyaml after https://github.com/smallfish/simpleyaml/pull/1 is merged
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCursor(t *testing.T) {
	YAML := func(s string) map[interface{}]interface{} {
		y, err := simpleyaml.NewYaml([]byte(s))
		So(err, ShouldBeNil)

		data, err := y.Map()
		So(err, ShouldBeNil)

		return data
	}

	Convey("Cursor", t, func() {
		Convey("parsing", func() {
			cursorIs := func(c *Cursor, nodes ...string) {
				So(c, ShouldNotBeNil)
				So(len(c.Nodes), ShouldEqual, len(nodes))

				for i := range nodes {
					So(c.Nodes[i], ShouldEqual, nodes[i])
				}
			}

			Convey("handles dotted-notation", func() {
				c, err := ParseCursor("x.y.z")
				So(err, ShouldBeNil)
				cursorIs(c, "x", "y", "z")
			})

			Convey("ignores the '$' sigil in the initial position", func() {
				c, err := ParseCursor("$.x.y.z")
				So(err, ShouldBeNil)
				cursorIs(c, "x", "y", "z")
			})

			Convey("handles the '$' sigil in any other position", func() {
				c, err := ParseCursor("x.$.y.z")
				So(err, ShouldBeNil)
				cursorIs(c, "x", "$", "y", "z")
			})

			Convey("handles traditional bracketed notation", func() {
				c, err := ParseCursor("nodes[1].sub")
				So(err, ShouldBeNil)
				cursorIs(c, "nodes", "1", "sub")
			})

			Convey("handles dotted bracketed notation", func() {
				c, err := ParseCursor("nodes.[1].sub")
				So(err, ShouldBeNil)
				cursorIs(c, "nodes", "1", "sub")
			})

			Convey("handles stacked bracketed notation", func() {
				c, err := ParseCursor("nodes[1][2][3]sub")
				So(err, ShouldBeNil)
				cursorIs(c, "nodes", "1", "2", "3", "sub")
			})

			Convey("handles non-integers in brackets", func() {
				c, err := ParseCursor("a[b][c].[one].two.[three]")
				So(err, ShouldBeNil)
				cursorIs(c, "a", "b", "c", "one", "two", "three")
			})

			Convey("handles dots as literals inside of bracket notation", func() {
				c, err := ParseCursor("meta[some.dotted.property].value")
				So(err, ShouldBeNil)
				cursorIs(c, "meta", "some.dotted.property", "value")
			})

			Convey("throws errors for unexpected closing brackets", func() {
				_, err := ParseCursor("x0]")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "syntax error: unexpected ']' at position 2")
			})

			Convey("throws errors for unexpected opening brackets", func() {
				_, err := ParseCursor("aa[[0]]")
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "syntax error: unexpected '[' at position 3")
			})
		})

		Convey("canonicalization", func() {
			tree := YAML(`---
key:
  list:
    - name: first
      value: 1
    - name: second
      value: 2
`)

			Convey("handles named list indexing", func() {
				c, _ := ParseCursor("key.list.second.value")
				canon, err := c.Canonical(tree)
				So(err, ShouldBeNil)
				So(canon, ShouldNotBeNil)
				So(canon.String(), ShouldEqual, "key.list.1.value")
			})

			Convey("handles maps all the way down", func() {
				c, _ := ParseCursor("key.list")
				canon, err := c.Canonical(tree)
				So(err, ShouldBeNil)
				So(canon, ShouldNotBeNil)
				So(canon.String(), ShouldEqual, "key.list")
			})

			Convey("handles missing keys", func() {
				c, _ := ParseCursor("key.that.does.not.exist")
				canon, err := c.Canonical(tree)
				So(err, ShouldNotBeNil)
				So(canon, ShouldBeNil)
			})

			Convey("handles missing list indexes", func() {
				c, _ := ParseCursor("key.list.42.vaule")
				canon, err := c.Canonical(tree)
				So(err, ShouldNotBeNil)
				So(canon, ShouldBeNil)
			})

			Convey("handles map-based addressing into a list", func() {
				c, _ := ParseCursor("key.list.xyzzy")
				canon, err := c.Canonical(tree)
				So(err, ShouldNotBeNil)
				So(canon, ShouldBeNil)
			})
		})

		Convey("manipulation", func() {
			c, _ := ParseCursor("x.y.z")

			Convey("Pop() removes an item from the end of the cursor", func() {
				c.Pop()
				So(c.String(), ShouldEqual, "x.y")
			})

			Convey("Pop() doesn't fail if you try to pop more nodes than are present", func() {
				for i := 0; i < 10; i++ {
					c.Pop()
				}
				So(c.String(), ShouldEqual, "")
			})

			Convey("Push() adds an item to the end of the cursor", func() {
				c.Push("omega")
				So(c.String(), ShouldEqual, "x.y.z.omega")
			})

			Convey("Component(n) retrieves arbitrary components", func() {
				So(c.Component(-1), ShouldEqual, "z")
				So(c.Component(-2), ShouldEqual, "y")
				So(c.Component(-3), ShouldEqual, "x")

				// out of range
				So(c.Component(-4), ShouldEqual, "")
				So(c.Component(-1023), ShouldEqual, "")
				So(c.Component(0), ShouldEqual, "")
			})

			Convey("Parent() is really a shorthand for Component(-2)", func() {
				So(c.Parent(), ShouldEqual, c.Component(-2))
			})
		})

		Convey("comparison via prefix match", func() {
			Convey("handles equivalent paths", func() {
				a, e1 := ParseCursor("x.y.z")
				b, e2 := ParseCursor("x.y.z")

				So(e1, ShouldBeNil)
				So(e2, ShouldBeNil)

				So(a.Under(b), ShouldBeFalse)
				So(b.Under(a), ShouldBeFalse)
			})

			Convey("handle disparate paths", func() {
				a, e1 := ParseCursor("A.B.C")
				b, e2 := ParseCursor("w.x.y.z")

				So(e1, ShouldBeNil)
				So(e2, ShouldBeNil)

				So(a.Under(b), ShouldBeFalse)
				So(b.Under(a), ShouldBeFalse)
			})

			Convey("handle empty paths", func() {
				a, e1 := ParseCursor("")
				b, e2 := ParseCursor("a.b.c")

				So(e1, ShouldBeNil)
				So(e2, ShouldBeNil)

				So(a.Under(b), ShouldBeFalse)
				So(b.Under(a), ShouldBeFalse)
			})

			Convey("correctly identifies subtree relationships", func() {
				a, e1 := ParseCursor("meta.templates")
				b, e2 := ParseCursor("meta.templates.0.job")

				So(e1, ShouldBeNil)
				So(e2, ShouldBeNil)

				So(a.Under(b), ShouldBeFalse)
				So(b.Under(a), ShouldBeTrue)
			})
		})

		Convey("resolving", func() {
			tree := YAML(
				`key:
  subkey:
    value: found it
  list:
    - name: apples
      quantity: 2
    - name: bananas
      quantity: 7
    - name: grapes
      quantity: 22
`)

			Convey("handles looking up the empty cursor (root)", func() {
				c, err := ParseCursor("")
				So(err, ShouldBeNil)

				v, err := c.Resolve(tree)
				So(err, ShouldBeNil)
				So(v, ShouldNotBeNil)
			})

			Convey("handles looking up using only map keys", func() {
				c, err := ParseCursor("key.subkey.value")
				So(err, ShouldBeNil)

				v, err := c.Resolve(tree)
				So(err, ShouldBeNil)
				So(v, ShouldEqual, "found it")
			})

			Convey("handles non-existent leaf/interior keys", func() {
				c, err := ParseCursor("key.subkey.enoent")
				So(err, ShouldBeNil)

				v, err := c.Resolve(tree)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "could not be found")
				So(v, ShouldBeNil)
			})

			Convey("handles non-existent root keys", func() {
				c, err := ParseCursor("not-a-key.sub.key.children")
				So(err, ShouldBeNil)

				v, err := c.Resolve(tree)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "could not be found")
				So(v, ShouldBeNil)
			})

			Convey("handles numeric list indices", func() {
				c, err := ParseCursor("key.list.0.name")
				So(err, ShouldBeNil)

				v, err := c.Resolve(tree)
				So(err, ShouldBeNil)
				So(v, ShouldEqual, "apples")
			})

			Convey("throws errors for out-of-bound indices", func() {
				c, err := ParseCursor("key.list.10.name")
				So(err, ShouldBeNil)

				v, err := c.Resolve(tree)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "could not be found")
				So(v, ShouldBeNil)
			})

			Convey("handles named list indices", func() {
				c, err := ParseCursor("key.list.apples.quantity")
				So(err, ShouldBeNil)

				v, err := c.Resolve(tree)
				So(err, ShouldBeNil)
				So(v, ShouldEqual, 2)
			})

			Convey("handles named list indices that don't exist", func() {
				c, err := ParseCursor("key.list.pomegranates.quantity")
				So(err, ShouldBeNil)

				v, err := c.Resolve(tree)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "could not be found")
				So(v, ShouldBeNil)
			})

			Convey("handles single glob position", func() {
				c, err := ParseCursor("key.list.*.name")
				So(err, ShouldBeNil)

				l, err := c.Glob(tree)
				So(err, ShouldBeNil)
				So(len(l), ShouldEqual, 3)
				So(l[0].String(), ShouldEqual, "key.list.0.name")
				So(l[1].String(), ShouldEqual, "key.list.1.name")
				So(l[2].String(), ShouldEqual, "key.list.2.name")
			})
		})
	})
}
