package spruce

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/starkandwayne/goutils/tree"
)

var _ = Describe("NullOperator", func() {
	Describe("Setup", func() {
		It("returns nil (no-op setup)", func() {
			op := NullOperator{Missing: "unknown-op"}
			Expect(op.Setup()).To(BeNil())
		})
	})

	Describe("Dependencies", func() {
		It("returns nil regardless of inputs", func() {
			op := NullOperator{Missing: "unknown-op"}
			deps := op.Dependencies(nil, nil, nil, []*tree.Cursor{{}})
			Expect(deps).To(BeNil())
		})
	})

	Describe("Run", func() {
		It("returns an error mentioning the missing operator name", func() {
			op := NullOperator{Missing: "frobulate"}
			ev := &Evaluator{Tree: map[interface{}]interface{}{}}
			resp, err := op.Run(ev, nil)
			Expect(resp).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("frobulate"))
			Expect(err.Error()).To(ContainSubstring("operator not defined"))
		})
	})
})
