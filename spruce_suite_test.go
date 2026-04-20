package spruce_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSpruce(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spruce Suite")
}
