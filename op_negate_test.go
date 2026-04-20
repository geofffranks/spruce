package spruce

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NegateOperator", func() {
	It("negates a boolean true to false", func() {
		ev := &Evaluator{Tree: evalYAML("val: true\nresult: (( negate val ))")}
		err := ev.RunPhase(EvalPhase)
		Expect(err).NotTo(HaveOccurred())
		Expect(ev.Tree["result"]).To(BeFalse())
	})

	It("negates a boolean false to true", func() {
		ev := &Evaluator{Tree: evalYAML("val: false\nresult: (( negate val ))")}
		err := ev.RunPhase(EvalPhase)
		Expect(err).NotTo(HaveOccurred())
		Expect(ev.Tree["result"]).To(BeTrue())
	})

	It("returns an error when given a non-boolean reference", func() {
		ev := &Evaluator{Tree: evalYAML("val: hello\nresult: (( negate val ))")}
		err := ev.RunPhase(EvalPhase)
		Expect(err).To(HaveOccurred())
	})
})
