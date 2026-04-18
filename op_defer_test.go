package spruce

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeferOperator", func() {
	It("defers evaluation by replacing with re-wrapped operator call", func() {
		ev := &Evaluator{Tree: evalYAML("val: (( defer grab meta.foo ))")}
		err := ev.RunPhase(EvalPhase)
		Expect(err).NotTo(HaveOccurred())
		Expect(ev.Tree["val"]).To(Equal("(( grab meta.foo ))"))
	})

	It("returns an error when defer has no arguments", func() {
		ev := &Evaluator{Tree: evalYAML("val: (( defer ))")}
		err := ev.RunPhase(EvalPhase)
		Expect(err).To(HaveOccurred())
	})
})
