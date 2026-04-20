package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var sprucePath string

func TestSpruceCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spruce CLI Suite")
}

var _ = BeforeSuite(func() {
	var err error
	sprucePath, err = gexec.Build("github.com/geofffranks/spruce/cmd/spruce")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
