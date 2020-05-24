package botta_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestBotta(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "botta Test Suite")
}
