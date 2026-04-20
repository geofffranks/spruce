package spruce

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadOperator", func() {
	It("loads a YAML file and returns a map", func() {
		// Get the absolute path to the test asset so this test is cwd-independent
		assetPath, err := filepath.Abs("assets/merge/first.yml")
		Expect(err).NotTo(HaveOccurred())

		ev := &Evaluator{Tree: evalYAML("result: (( load \"" + assetPath + "\" ))")}
		runErr := ev.RunPhase(EvalPhase)
		Expect(runErr).NotTo(HaveOccurred())
		Expect(ev.Tree["result"]).NotTo(BeNil())
		result, ok := ev.Tree["result"].(map[interface{}]interface{})
		Expect(ok).To(BeTrue())
		Expect(result).To(HaveKey("key"))
	})

	It("loads a YAML file via SPRUCE_FILE_BASE_PATH", func() {
		assetDir, err := filepath.Abs("assets/merge")
		Expect(err).NotTo(HaveOccurred())
		os.Setenv("SPRUCE_FILE_BASE_PATH", assetDir)
		defer os.Unsetenv("SPRUCE_FILE_BASE_PATH")

		ev := &Evaluator{Tree: evalYAML(`result: (( load "first.yml" ))`)}
		runErr := ev.RunPhase(EvalPhase)
		Expect(runErr).NotTo(HaveOccurred())
		Expect(ev.Tree["result"]).NotTo(BeNil())
	})

	It("returns an error for a missing file", func() {
		ev := &Evaluator{Tree: evalYAML(`result: (( load "nonexistent_file_xyz.yml" ))`)}
		err := ev.RunPhase(EvalPhase)
		Expect(err).To(HaveOccurred())
	})
})
