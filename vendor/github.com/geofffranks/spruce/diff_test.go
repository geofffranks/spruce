package spruce

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"testing"
	"unicode"

	// Use geofffranks forks to persist the fix in https://github.com/go-yaml/yaml/pull/133/commits
	// Also https://github.com/go-yaml/yaml/pull/195
	"github.com/geofffranks/simpleyaml"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDiff(t *testing.T) {
	YAML := func(s string) map[interface{}]interface{} {
		y, err := simpleyaml.NewYaml([]byte(s))
		So(err, ShouldBeNil)

		data, err := y.Map()
		So(err, ShouldBeNil)

		return data
	}

	trim := func(s string) string {
		/* find the shortest prefix of space characters */
		shortest := regexp.MustCompile(`^ +`).FindString(s)
		s = strings.TrimRightFunc(s, unicode.IsSpace)
		return regexp.MustCompile(`(?m)^`+shortest).ReplaceAllString(s, "")
	}

	var test, a, b, diff string
	var current *string
	runtest := func() {
		if test == "" {
			return
		}
		Convey(test, t, func() {
			d, err := Diff(YAML(a), YAML(b))
			So(err, ShouldBeNil)
			So(trim(d.String("$")), ShouldEqual, trim(diff))
		})
	}

	testPat := regexp.MustCompile(`^###+\s+(.*)\s*$`)
	in, err := os.Open("tests/diff")
	if err != nil {
		panic(err)
	}

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
}
