package spruce

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/geofffranks/simpleyaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Diff", func() {
	YAML := func(s string) map[interface{}]interface{} {
		y, err := simpleyaml.NewYaml([]byte(s))
		Expect(err).NotTo(HaveOccurred())

		data, err := y.Map()
		Expect(err).NotTo(HaveOccurred())

		return data
	}

	trim := func(s string) string {
		shortest := regexp.MustCompile(`^ +`).FindString(s)
		s = strings.TrimRightFunc(s, unicode.IsSpace)
		return regexp.MustCompile(`(?m)^`+shortest).ReplaceAllString(s, "")
	}

	It("passes all diff test cases from tests/diff", func() {
		var test, a, b, diff string
		var current *string
		testPat := regexp.MustCompile(`^###+\s+(.*)\s*$`)

		runtest := func() {
			if test == "" {
				return
			}
			By(test)
			d, err := Diff(YAML(a), YAML(b))
			Expect(err).NotTo(HaveOccurred())
			Expect(trim(d.String("$"))).To(Equal(trim(diff)))
		}

		in, err := os.Open("tests/diff")
		Expect(err).NotTo(HaveOccurred())
		defer in.Close()

		s := bufio.NewScanner(in)
		for s.Scan() {
			if testPat.MatchString(s.Text()) {
				m := testPat.FindStringSubmatch(s.Text())
				runtest()
				test, a, b, diff = m[1], "", "", ""
				continue
			}

			if s.Text() == "---" {
				if a == "" {
					current = &a
				} else if b == "" {
					current = &b
				} else {
					current = &diff
				}
				continue
			}

			if current != nil {
				*current = *current + s.Text() + "\n"
			}
		}
		runtest()
	})
})
