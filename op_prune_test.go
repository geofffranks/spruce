package spruce

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PruneOperator", func() {
	It("marks paths for pruning during evaluation", func() {
		ev := &Evaluator{Tree: evalYAML("meta:\n  foo: bar\nremove: (( prune ))\nresult: (( grab meta.foo ))")}
		err := ev.Run(nil, nil)
		Expect(err).NotTo(HaveOccurred())
		_, exists := ev.Tree["remove"]
		Expect(exists).To(BeFalse())
	})

	It("keeps other keys after pruning", func() {
		ev := &Evaluator{Tree: evalYAML("keep: hello\nremove: (( prune ))")}
		err := ev.Run(nil, nil)
		Expect(err).NotTo(HaveOccurred())
		_, exists := ev.Tree["remove"]
		Expect(exists).To(BeFalse())
		Expect(ev.Tree["keep"]).To(Equal("hello"))
	})
})
