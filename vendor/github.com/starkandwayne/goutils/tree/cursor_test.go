package tree_test

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"gopkg.in/yaml.v2"

	"github.com/starkandwayne/goutils/tree"
)

type cursorMatcher struct {
	nodes []string
}

func BeCursor(nodes ...string) types.GomegaMatcher {
	return &cursorMatcher{nodes: nodes}
}

func (m *cursorMatcher) Match(actual interface{}) (bool, error) {
	cursor, ok := actual.(*tree.Cursor)
	if !ok {
		return false, fmt.Errorf("BeCursor matcher expects a cursor")
	}

	if len(cursor.Nodes) != len(m.nodes) {
		return false, nil
	}
	for i := range cursor.Nodes {
		if cursor.Nodes[i] != m.nodes[i] {
			return false, nil
		}
	}

	return true, nil
}

func (m *cursorMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%s\nto equal\n\t%s", actual.(tree.Cursor), strings.Join(m.nodes, "."))
}

func (m *cursorMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%s\nto not equal\n\t%s", actual.(tree.Cursor), strings.Join(m.nodes, "."))
}

var _ = Describe("Cursors", func() {
	Context("Basic Cursor Parsing", func() {
		It("handles dotted-notation", func() {
			c, err := tree.ParseCursor("x.y.z")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(c).Should(BeCursor("x", "y", "z"))
		})

		It("ignores the '$' sigil in the initial position", func() {
			c, err := tree.ParseCursor("$.x.y.z")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(c).Should(BeCursor("x", "y", "z"))
		})

		It("handles the '$' sigil in any other position", func() {
			c, err := tree.ParseCursor("x.$.y.z")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(c).Should(BeCursor("x", "$", "y", "z"))
		})

		It("handles traditional bracketed notation", func() {
			c, err := tree.ParseCursor("nodes[1].sub")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(c).Should(BeCursor("nodes", "1", "sub"))
		})

		It("handles dotted bracketed notation", func() {
			c, err := tree.ParseCursor("nodes.[1].sub")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(c).Should(BeCursor("nodes", "1", "sub"))
		})

		It("handles stacked bracketed notation", func() {
			c, err := tree.ParseCursor("nodes[1][2][3]sub")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(c).Should(BeCursor("nodes", "1", "2", "3", "sub"))
		})

		It("handles non-integers in brackets", func() {
			c, err := tree.ParseCursor("a[b][c].[one].two.[three]")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(c).Should(BeCursor("a", "b", "c", "one", "two", "three"))
		})

		It("handles dots as literals inside of bracket notation", func() {
			c, err := tree.ParseCursor("meta[some.dotted.property].value")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(c).Should(BeCursor("meta", "some.dotted.property", "value"))
		})

		It("throws errors for unexpected closing brackets", func() {
			_, err := tree.ParseCursor("x0]")
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("syntax error: unexpected ']' at position 2"))
		})

		It("throws errors for unexpected opening brackets", func() {
			_, err := tree.ParseCursor("aa[[0]]")
			Ω(err).Should(HaveOccurred())
			Ω(err.Error()).Should(Equal("syntax error: unexpected '[' at position 3"))
		})
	})

	Context("Canonical()", func() {
		var T interface{}

		BeforeEach(func() {
			s := `
{
  "key":{
    "list":[
      {
        "name" : "first",
        "value": 1
      },{
        "name" : "second",
        "value": 2
      }
    ]
  }
}`
			Ω(json.Unmarshal([]byte(s), &T)).Should(Succeed())
		})

		It("handles named list indexing", func() {
			c, _ := tree.ParseCursor("key.list.second.value")
			canon, err := c.Canonical(T)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(canon).ShouldNot(BeNil())
			Ω(canon.String()).Should(Equal("key.list.1.value"))
		})
		It("handles regular list indexing", func() {
			c, _ := tree.ParseCursor("key.list.1.value")
			canon, err := c.Canonical(T)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(canon).ShouldNot(BeNil())
			Ω(canon.String()).Should(Equal("key.list.1.value"))
		})

		It("handles maps all the way down", func() {
			c, _ := tree.ParseCursor("key.list")
			canon, err := c.Canonical(T)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(canon).ShouldNot(BeNil())
			Ω(canon.String()).Should(Equal("key.list"))
		})

		It("handles missing keys", func() {
			c, _ := tree.ParseCursor("key.that.does.not.exist")
			canon, err := c.Canonical(T)
			Ω(err).Should(HaveOccurred())
			Ω(canon).Should(BeNil())
		})

		It("handles missing list indexes", func() {
			c, _ := tree.ParseCursor("key.list.42.vaule")
			canon, err := c.Canonical(T)
			Ω(err).Should(HaveOccurred())
			Ω(canon).Should(BeNil())
		})

		It("handles map-based addressing into a list", func() {
			c, _ := tree.ParseCursor("key.list.xyzzy")
			canon, err := c.Canonical(T)
			Ω(err).Should(HaveOccurred())
			Ω(canon).Should(BeNil())
		})
	})

	Context("Pop()", func() {
		c, _ := tree.ParseCursor("x.y.z")

		It("removes an item from the end of the cursor", func() {
			c.Pop()
			Ω(c.String()).Should(Equal("x.y"))
		})

		It("doesn't fail if you try to pop more nodes than are present", func() {
			for i := 0; i < 10; i++ {
				c.Pop()
			}
			Ω(c.String()).Should(Equal(""))
		})
	})

	Context("Push()", func() {
		c, _ := tree.ParseCursor("x.y.z")
		It("adds an item to the end of the cursor", func() {
			c.Push("omega")
			Ω(c.String()).Should(Equal("x.y.z.omega"))
		})
	})

	Context("Component()", func() {
		c, _ := tree.ParseCursor("x.y.z")
		It("retrieves arbitrary components", func() {
			Ω(c.Component(-1)).Should(Equal("z"))
			Ω(c.Component(-2)).Should(Equal("y"))
			Ω(c.Component(-3)).Should(Equal("x"))

			// out of range
			Ω(c.Component(-4)).Should(Equal(""))
			Ω(c.Component(-1023)).Should(Equal(""))
			Ω(c.Component(0)).Should(Equal(""))
		})
	})

	Context("Parent()", func() {
		c, _ := tree.ParseCursor("x.y.z")
		It("Parent() is really a shorthand for Component(-2)", func() {
			Ω(c.Parent()).Should(Equal(c.Component(-2)))
		})
	})

	Context("Under()", func() {
		It("handles equivalent paths", func() {
			a, e1 := tree.ParseCursor("x.y.z")
			b, e2 := tree.ParseCursor("x.y.z")

			Ω(e1).Should(BeNil())
			Ω(e2).Should(BeNil())

			Ω(a.Under(b)).Should(BeFalse())
			Ω(b.Under(a)).Should(BeFalse())
		})

		It("handles disparate paths", func() {
			a, e1 := tree.ParseCursor("A.B.C")
			b, e2 := tree.ParseCursor("w.x.y.z")

			Ω(e1).Should(BeNil())
			Ω(e2).Should(BeNil())

			Ω(a.Under(b)).Should(BeFalse())
			Ω(b.Under(a)).Should(BeFalse())
		})

		It("handles empty paths", func() {
			a, e1 := tree.ParseCursor("")
			b, e2 := tree.ParseCursor("a.b.c")

			Ω(e1).Should(BeNil())
			Ω(e2).Should(BeNil())

			Ω(a.Under(b)).Should(BeFalse())
			Ω(b.Under(a)).Should(BeFalse())
		})

		It("correctly identifies subtree relationships", func() {
			a, e1 := tree.ParseCursor("meta.templates")
			b, e2 := tree.ParseCursor("meta.templates.0.job")

			Ω(e1).Should(BeNil())
			Ω(e2).Should(BeNil())

			Ω(a.Under(b)).Should(BeFalse())
			Ω(b.Under(a)).Should(BeTrue())
		})
	})

	Context("Contains()", func() {
		It("handles equivalent paths", func() {
			a, e1 := tree.ParseCursor("x.y.z")
			b, e2 := tree.ParseCursor("x.y.z")

			Ω(e1).Should(BeNil())
			Ω(e2).Should(BeNil())

			Ω(a.Contains(b)).Should(BeTrue())
			Ω(b.Contains(a)).Should(BeTrue())
		})

		It("handles disparate paths", func() {
			a, e1 := tree.ParseCursor("A.B.C")
			b, e2 := tree.ParseCursor("w.x.y.z")

			Ω(e1).Should(BeNil())
			Ω(e2).Should(BeNil())

			Ω(a.Contains(b)).Should(BeFalse())
			Ω(b.Contains(a)).Should(BeFalse())
		})

		It("handles empty paths", func() {
			a, e1 := tree.ParseCursor("")
			b, e2 := tree.ParseCursor("a.b.c")

			Ω(e1).Should(BeNil())
			Ω(e2).Should(BeNil())

			Ω(a.Contains(b)).Should(BeFalse())
			Ω(b.Contains(a)).Should(BeFalse())
		})

		It("correctly identifies subtree relationships", func() {
			a, e1 := tree.ParseCursor("meta.templates")
			b, e2 := tree.ParseCursor("meta.templates.0.job")

			Ω(e1).Should(BeNil())
			Ω(e2).Should(BeNil())

			Ω(a.Contains(b)).Should(BeTrue())
			Ω(b.Contains(a)).Should(BeFalse())
		})
	})

	Context("Resolve()", func() {
		TestResolve := func(fn func() interface{}) {
			var T interface{}
			BeforeEach(func() {
				T = fn()
			})

			It("handles looking up the empty cursor (root)", func() {
				c, err := tree.ParseCursor("")
				Ω(err).ShouldNot(HaveOccurred())

				v, err := c.Resolve(T)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).ShouldNot(BeNil())
			})

			It("handles looking up using only map keys", func() {
				c, err := tree.ParseCursor("key.subkey.value")
				Ω(err).ShouldNot(HaveOccurred())

				v, err := c.Resolve(T)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal("found it"))
			})

			It("handles non-existent leaf/interior keys", func() {
				c, err := tree.ParseCursor("key.subkey.enoent")
				Ω(err).ShouldNot(HaveOccurred())

				v, err := c.Resolve(T)
				Ω(err).Should(HaveOccurred())
				Ω(err.Error()).Should(ContainSubstring("could not be found"))
				Ω(v).Should(BeNil())
			})

			It("handles non-existent root keys", func() {
				c, err := tree.ParseCursor("not-a-key.sub.key.children")
				Ω(err).ShouldNot(HaveOccurred())

				v, err := c.Resolve(T)
				Ω(err).Should(HaveOccurred())
				Ω(err.Error()).Should(ContainSubstring("could not be found"))
				Ω(v).Should(BeNil())
			})

			It("handles numeric list indices", func() {
				c, err := tree.ParseCursor("key.list.0.name")
				Ω(err).ShouldNot(HaveOccurred())

				v, err := c.Resolve(T)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal("apples"))
			})

			It("throws errors for out-of-bound indices", func() {
				c, err := tree.ParseCursor("key.list.10.name")
				Ω(err).ShouldNot(HaveOccurred())

				v, err := c.Resolve(T)
				Ω(err).Should(HaveOccurred())
				Ω(err.Error()).Should(ContainSubstring("could not be found"))
				Ω(v).Should(BeNil())
			})

			It("handles named list indices", func() {
				c, err := tree.ParseCursor("key.list.apples.quantity")
				Ω(err).ShouldNot(HaveOccurred())

				v, err := c.Resolve(T)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(BeNumerically("==", 2))
			})

			It("handles named list indices that don't exist", func() {
				c, err := tree.ParseCursor("key.list.pomegranates.quantity")
				Ω(err).ShouldNot(HaveOccurred())

				v, err := c.Resolve(T)
				Ω(err).Should(HaveOccurred())
				Ω(err.Error()).Should(ContainSubstring("could not be found"))
				Ω(v).Should(BeNil())
			})

			It("handles maps with leaf numeric indices", func() {
				c, err := tree.ParseCursor("key.nums.1")
				Ω(err).ShouldNot(HaveOccurred())

				v, err := c.Resolve(T)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal("one"))
			})

			It("handles maps with interior numeric indices", func() {
				c, err := tree.ParseCursor("key.nums.2.name")
				Ω(err).ShouldNot(HaveOccurred())

				v, err := c.Resolve(T)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(v).Should(Equal("two"))
			})

			It("handles single glob position", func() {
				c, err := tree.ParseCursor("key.list.*.name")
				Ω(err).ShouldNot(HaveOccurred())

				l, err := c.Glob(T)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(l)).Should(Equal(3))
				Ω(l[0].String()).Should(Equal("key.list.0.name"))
				Ω(l[1].String()).Should(Equal("key.list.1.name"))
				Ω(l[2].String()).Should(Equal("key.list.2.name"))
			})

			It("handles terminal glob position", func() {
				c, err := tree.ParseCursor("key.simple.list.*")
				Ω(err).ShouldNot(HaveOccurred())

				l, err := c.Glob(T)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(len(l)).Should(Equal(3))
				Ω(l[0].String()).Should(Equal("key.simple.list.0"))
				Ω(l[1].String()).Should(Equal("key.simple.list.1"))
				Ω(l[2].String()).Should(Equal("key.simple.list.2"))
			})

			It("doesn't error when globbing with '*' doesn't find things", func() {
				c, err := tree.ParseCursor("key.list.*.optional")
				Expect(err).ShouldNot(HaveOccurred())

				l, err := c.Glob(T)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(len(l)).Should(Equal(1))
				Expect(l[0].String()).Should(Equal("key.list.1.optional"))
			})
			It("throws an error when globbing without '*' doesn't find things", func() {
				c, err := tree.ParseCursor("key.list.0.optional")
				Expect(err).ShouldNot(HaveOccurred())

				l, err := c.Glob(T)
				Expect(err).Should(HaveOccurred())
				Expect(l).Should(BeNil())
			})
			It("throws an error when missing nodes are prior to the first '*' of a Glob", func() {
				c, err := tree.ParseCursor("key.listnotfound.*.missing")
				Expect(err).ShouldNot(HaveOccurred())

				l, err := c.Glob(T)
				Expect(err).Should(HaveOccurred())
				Expect(l).Should(BeNil())
			})
		}

		Context("With JSON trees", func() {
			TestResolve(func() (T interface{}) {
				s := `
{
  "key": {
    "subkey": {
      "value": "found it"
    },
    "list": [
      {
        "name": "apples",
        "quantity": 2
      },{
        "name": "bananas",
        "quantity": 7,
        "optional": "present"
      },{
        "name": "grapes",
        "quantity": 22
      }
    ],
    "nums": {
      "1": "one",
      "2": {
        "name": "two",
        "value": 2
      }
    },
    "simple": {
      "list": [
        "first",
        "second",
        "third"
      ]
    }
  }
}`
				Ω(json.Unmarshal([]byte(s), &T)).Should(Succeed())
				return
			})
		})

		Context("With YAML trees", func() {
			TestResolve(func() (T interface{}) {
				s := `---
key:
  subkey:
    value: found it
  list:
    - name:     apples
      quantity: 2
    - name:     bananas
      quantity: 7
      optional: present
    - name:     grapes
      quantity: 22
  nums:
    1: one
    2:
      name: two
      value: 2
  simple:
    list:
      - first
      - second
      - third
`
				Ω(yaml.Unmarshal([]byte(s), &T)).Should(Succeed())
				return
			})
		})
	})
})
