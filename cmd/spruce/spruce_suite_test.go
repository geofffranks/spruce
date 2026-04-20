package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/starkandwayne/goutils/ansi"
)

var sprucePath string

func TestSpruceCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spruce CLI Suite")
}

var _ = BeforeSuite(func() {
	// Disable ANSI colors in error messages so in-process assertions on
	// err.Error() (e.g. internal_test.go) are stable regardless of whether
	// stdout is a TTY. Subprocess (gexec) tests are unaffected: the spruce
	// binary decides color based on its own stdout, which is piped to gexec.
	ansi.Color(false)

	var err error
	sprucePath, err = gexec.Build("github.com/geofffranks/spruce/cmd/spruce")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
