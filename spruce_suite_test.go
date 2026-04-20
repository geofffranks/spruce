package spruce_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/starkandwayne/goutils/ansi"
)

func TestSpruce(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spruce Suite")
}

// Disable ANSI colors in error messages so string-equality assertions work
// regardless of whether the terminal running the test suite is a TTY. When
// stdout is a TTY (iTerm, Terminal.app, etc.) the ansi package emits color
// escapes that break Equal/ContainSubstring matchers on err.Error().
var _ = BeforeSuite(func() {
	ansi.Color(false)
})
