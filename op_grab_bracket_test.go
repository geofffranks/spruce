package spruce

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Grab Operator: dynamic bracket-notation lookup", func() {
	op := GrabOperator{}

	dynRef := func(s string) *Expr {
		return &Expr{Type: Reference, Reference: cursor(s), BracketedNodes: bracketsOf(s)}
	}

	It("can grab a value using a dynamic bracket reference", func() {
		ev := &Evaluator{
			Tree: opYAML(`key:
  subkey: found it
  other: value 2
lookup: subkey
`),
		}
		r, err := op.Run(ev, []*Expr{
			dynRef("key[lookup]"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("found it"))
	})

	It("can grab a value using a nested path in a bracket reference", func() {
		ev := &Evaluator{
			Tree: opYAML(`key:
  subkey: found it
meta:
  which: subkey
`),
		}
		r, err := op.Run(ev, []*Expr{
			dynRef("key[meta.which]"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("found it"))
	})

	It("returns error when the bracket key reference resolves to a non-scalar", func() {
		ev := &Evaluator{
			Tree: opYAML(`key:
  subkey: found it
lookup:
  nested: subkey
`),
		}
		_, err := op.Run(ev, []*Expr{
			dynRef("key[lookup]"),
		})
		Expect(err).To(HaveOccurred())
	})

	It("can grab a value using a numeric-valued bracket reference", func() {
		ev := &Evaluator{
			Tree: opYAML(`versions:
  "1": first
  "2": second
pick: 2
`),
		}
		r, err := op.Run(ev, []*Expr{
			dynRef("versions[pick]"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("second"))
	})

	It("can grab a value using an environment variable in a bracket reference", func() {
		os.Setenv("BRACKET_SUB_KEY", "subkey")
		defer os.Unsetenv("BRACKET_SUB_KEY")

		ev := &Evaluator{
			Tree: opYAML(`key:
  subkey: found it
  other: value 2
`),
		}
		r, err := op.Run(ev, []*Expr{
			dynRef("key[$BRACKET_SUB_KEY]"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("found it"))
	})
})
