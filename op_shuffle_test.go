package spruce

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ShuffleOperator", func() {
	It("returns a shuffled list of the same length", func() {
		ev := &Evaluator{Tree: evalYAML("list:\n- a\n- b\n- c\nresult: (( shuffle list ))")}
		err := ev.RunPhase(EvalPhase)
		Expect(err).NotTo(HaveOccurred())
		result, ok := ev.Tree["result"].([]interface{})
		Expect(ok).To(BeTrue())
		Expect(result).To(HaveLen(3))
		Expect(result).To(ContainElements("a", "b", "c"))
	})

	It("returns an error when given a map reference", func() {
		ev := &Evaluator{Tree: evalYAML("mymap:\n  key: value\nresult: (( shuffle mymap ))")}
		err := ev.RunPhase(EvalPhase)
		Expect(err).To(HaveOccurred())
	})
})
