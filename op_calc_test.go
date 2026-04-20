package spruce

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CalcOperator", func() {
	It("evaluates simple arithmetic", func() {
		ev := &Evaluator{Tree: evalYAML(`result: (( calc "2 + 3" ))`)}
		err := ev.RunPhase(EvalPhase)
		Expect(err).NotTo(HaveOccurred())
		Expect(ev.Tree["result"]).To(BeNumerically("==", 5))
	})

	It("evaluates with dotted-path references", func() {
		// calc operator resolves references via dot-notation paths (e.g. meta.a)
		ev := &Evaluator{Tree: evalYAML("meta:\n  a: 10\n  b: 20\nresult: (( calc \"meta.a + meta.b\" ))")}
		err := ev.RunPhase(EvalPhase)
		Expect(err).NotTo(HaveOccurred())
		Expect(ev.Tree["result"]).To(BeNumerically("==", 30))
	})

	It("returns an error for non-literal arguments", func() {
		ev := &Evaluator{Tree: evalYAML("a: 10\nresult: (( calc a ))")}
		err := ev.RunPhase(EvalPhase)
		Expect(err).To(HaveOccurred())
	})
})
