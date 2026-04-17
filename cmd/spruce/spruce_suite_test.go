package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSpruceCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spruce CLI Suite")
}
