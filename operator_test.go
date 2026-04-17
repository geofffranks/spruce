package spruce

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"github.com/aws/aws-sdk-go/aws"         //nolint:staticcheck // SA1019: aws-sdk-go v1 deprecated; v2 migration tracked separately
	"github.com/aws/aws-sdk-go/service/secretsmanager" //nolint:staticcheck // SA1019: aws-sdk-go v1 deprecated; v2 migration tracked separately
	"github.com/aws/aws-sdk-go/service/ssm"             //nolint:staticcheck // SA1019: aws-sdk-go v1 deprecated; v2 migration tracked separately

	"github.com/geofffranks/simpleyaml"
	"github.com/geofffranks/spruce/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/starkandwayne/goutils/tree"
)

// Helper functions shared across all operator Describe blocks.

var cursor = func(s string) *tree.Cursor {
	c, err := tree.ParseCursor(s)
	Expect(err).NotTo(HaveOccurred())
	return c
}

var opYAML = func(s string) map[interface{}]interface{} {
	y, err := simpleyaml.NewYaml([]byte(s))
	Expect(err).NotTo(HaveOccurred())
	data, err := y.Map()
	Expect(err).NotTo(HaveOccurred())
	return data
}

var ref = func(s string) *Expr {
	return &Expr{Type: Reference, Reference: cursor(s)}
}
var str = func(s string) *Expr {
	return &Expr{Type: Literal, Literal: s}
}
var num = func(v int64) *Expr {
	return &Expr{Type: Literal, Literal: v}
}
var null = func() *Expr {
	return &Expr{Type: Literal, Literal: nil}
}
var env = func(s string) *Expr {
	return &Expr{Type: EnvVar, Name: s}
}
var or = func(l *Expr, r *Expr) *Expr {
	return &Expr{Type: LogicalOr, Left: l, Right: r}
}

var exprOk func(*Expr, *Expr)

func init() {
	exprOk = func(got *Expr, want *Expr) {
		Expect(got).NotTo(BeNil())
		Expect(want).NotTo(BeNil())
		Expect(got.Type).To(Equal(want.Type))
		switch want.Type {
		case Literal:
			if want.Literal == nil {
				Expect(got.Literal).To(BeNil())
			} else {
				Expect(got.Literal).To(Equal(want.Literal))
			}
		case Reference:
			Expect(got.Reference.String()).To(Equal(want.Reference.String()))
		case LogicalOr:
			exprOk(got.Left, want.Left)
			exprOk(got.Right, want.Right)
		}
	}
}

var _ = Describe("Parser", func() {
	Describe("parses op calls in their entirety", func() {
		phase := EvalPhase

		opOk := func(code string, name string, args ...*Expr) {
			op, err := ParseOpcall(phase, code)
			Expect(err).NotTo(HaveOccurred())
			Expect(op).NotTo(BeNil())

			_, ok := op.op.(NullOperator)
			Expect(ok).To(BeTrue())
			Expect(op.op.(NullOperator).Missing).To(Equal(name))

			Expect(len(op.args)).To(Equal(len(args)))
			for i, expect := range args {
				exprOk(op.args[i], expect)
			}
		}

		opErr := func(code string, msg string) {
			_, err := ParseOpcall(phase, code)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(msg))
		}

		opIgnore := func(code string) {
			op, err := ParseOpcall(phase, code)
			Expect(op).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		}

		It("handles opcodes with and without arguments", func() {
			opOk(`(( null 42 ))`, "null", num(42))
			opOk(`(( null 1 2 3 4 ))`, "null", num(1), num(2), num(3), num(4))
		})

		It("ignores optional whitespace", func() {
			args := []*Expr{num(1), num(2), num(3)}
			opOk(`((null 1 2 3))`, "null", args...)
			opOk(`((null 1	2	3))`, "null", args...)
			opOk(`((null 1	2	3	))`, "null", args...)
			opOk(`((null 1 	 2 	 3 	 ))`, "null", args...)
		})

		It("allows use of commas to separate arguments", func() {
			args := []*Expr{num(1), num(2), num(3)}
			opOk(`((null 1, 2, 3))`, "null", args...)
			opOk(`((null 1,	2,	3))`, "null", args...)
			opOk(`((null 1,	2,	3,	))`, "null", args...)
			opOk(`((null 1 ,	 2 ,	 3 ,	 ))`, "null", args...)
		})

		It("allows use of parentheses around arguments", func() {
			args := []*Expr{num(1), num(2), num(3)}
			opOk(`((null(1,2,3)))`, "null", args...)
			opOk(`((null(1, 2, 3) ))`, "null", args...)
			opOk(`((null( 1,	2,	3)))`, "null", args...)
			opOk(`((null (1,	2,	3)	))`, "null", args...)
			opOk(`((null (1 ,	 2 ,	 3)	 ))`, "null", args...)
		})

		It("handles string literal arguments", func() {
			opOk(`(( null "string" ))`, "null", str("string"))
			opOk(`(( null "string with whitespace" ))`, "null", str("string with whitespace"))
			opOk(`(( null "a \"quoted\" string" ))`, "null", str(`a "quoted" string`))
			opOk(`(( null "\\escaped" ))`, "null", str(`\escaped`))
		})

		It("handles reference (cursor) arguments", func() {
			opOk(`(( null x.y.z ))`, "null", ref("x.y.z"))
			opOk(`(( null x.[0].z ))`, "null", ref("x.0.z"))
			opOk(`(( null x[0].z ))`, "null", ref("x.0.z"))
			opOk(`(( null x[0]z ))`, "null", ref("x.0.z"))
		})

		It("handles mixed collections of argument types", func() {
			opOk(`(( xyzzy "string" x.y.z 42  ))`, "xyzzy", str("string"), ref("x.y.z"), num(42))
			opOk(`(( xyzzy("string" x.y.z 42) ))`, "xyzzy", str("string"), ref("x.y.z"), num(42))
		})

		It("handles expression-based operands", func() {
			opOk(`(( null meta.key || "default" ))`, "null",
				or(ref("meta.key"), str("default")))

			opOk(`(( null meta.key || "default" "second" ))`, "null",
				or(ref("meta.key"), str("default")),
				str("second"))

			opOk(`(( null meta.key || "default", "second" ))`, "null",
				or(ref("meta.key"), str("default")),
				str("second"))

			opOk(`(( null meta.key || "default", meta.other || nil ))`, "null",
				or(ref("meta.key"), str("default")),
				or(ref("meta.other"), null()))

			opOk(`(( null meta.key || "default"     meta.other || nil ))`, "null",
				or(ref("meta.key"), str("default")),
				or(ref("meta.other"), null()))
		})

		It("handles environment variables as operands", func() {
			os.Setenv("SPRUCE_FOO", "first test")
			os.Setenv("_SPRUCE", "_sprucify!")
			os.Setenv("ENOENT", "")
			os.Setenv("http_proxy", "no://thank/you")
			os.Setenv("variable.with.dots", "dots are ok")

			opOk(`(( null $SPRUCE_FOO ))`, "null", env("SPRUCE_FOO"))
			opOk(`(( null $_SPRUCE ))`, "null", env("_SPRUCE"))
			opOk(`(( null $ENOENT || $SPRUCE_FOO ))`, "null",
				or(env("ENOENT"), env("SPRUCE_FOO")))
			opOk(`(( null $http_proxy))`, "null", env("http_proxy"))
			opOk(`(( null $variable.with.dots ))`, "null", env("variable.with.dots"))
		})

		It("throws errors for malformed expression", func() {
			opErr(`(( null meta.key ||, nil ))`,
				`syntax error near: meta.key ||, nil`)

			opErr(`(( null || ))`,
				`syntax error near: ||`)

			opErr(`(( null || meta.key ))`,
				`syntax error near: || meta.key`)

			opErr(`(( null meta.key || || ))`,
				`syntax error near: meta.key || ||`)
		})

		It("ignores spiff-like bang-notation", func() {
			opIgnore(`((!credhub))`)
		})

		It("ignores BOSH varnames that aren't null-arity operators", func() {
			opIgnore(`((var-name))`)
		})
	})
})

var _ = Describe("Expression Engine", func() {
	var e *Expr
	var testTree map[interface{}]interface{}

	evaluate := func(e *Expr, tree map[interface{}]interface{}) interface{} {
		v, err := e.Evaluate(tree)
		Expect(err).NotTo(HaveOccurred())
		return v
	}

	It("Literals evaluate to themselves", func() {
		e = &Expr{Type: Literal, Literal: "value"}
		Expect(evaluate(e, testTree)).To(Equal("value"))

		e = &Expr{Type: Literal, Literal: ""}
		Expect(evaluate(e, testTree)).To(Equal(""))

		e = &Expr{Type: Literal, Literal: nil}
		Expect(evaluate(e, testTree)).To(BeNil())
	})

	It("References evaluate to the referenced part of the YAML tree", func() {
		testTree = opYAML(`---
meta:
  foo: FOO
  bar: BAR
`)

		e = &Expr{Type: Reference, Reference: cursor("meta.foo")}
		Expect(evaluate(e, testTree)).To(Equal("FOO"))

		e = &Expr{Type: Reference, Reference: cursor("meta.bar")}
		Expect(evaluate(e, testTree)).To(Equal("BAR"))
	})

	It("|| operator evaluates to the first found value", func() {
		testTree = opYAML(`---
meta:
  foo: FOO
  bar: BAR
`)

		Expect(evaluate(or(str("first"), str("second")), testTree)).To(Equal("first"))
		Expect(evaluate(or(ref("meta.foo"), str("second")), testTree)).To(Equal("FOO"))
		Expect(evaluate(or(ref("meta.ENOENT"), ref("meta.foo")), testTree)).To(Equal("FOO"))
	})

	It("|| operator treats nil as a found value", func() {
		testTree = opYAML(`---
meta:
  foo: FOO
  bar: BAR
`)

		Expect(evaluate(or(null(), str("second")), testTree)).To(BeNil())
		Expect(evaluate(or(ref("meta.ENOENT"), null()), testTree)).To(BeNil())
	})
})

var _ = Describe("Expression Reduction Algorithm", func() {
	var orig, final *Expr
	var err error

	It("ignores singleton expression", func() {
		orig = str("string")
		final, err = orig.Reduce()
		Expect(err).NotTo(HaveOccurred())
		exprOk(final, orig)

		orig = null()
		final, err = orig.Reduce()
		Expect(err).NotTo(HaveOccurred())
		exprOk(final, orig)

		orig = ref("meta.key")
		final, err = orig.Reduce()
		Expect(err).NotTo(HaveOccurred())
		exprOk(final, orig)
	})

	It("handles normal alternates that terminated in a literal", func() {
		orig = or(ref("a.b.c"), str("default"))
		final, err = orig.Reduce()
		Expect(err).NotTo(HaveOccurred())
		exprOk(final, orig)
	})

	It("throws errors (warnings) for unreachable alternates", func() {
		orig = or(null(), str("ignored"))
		final, err = orig.Reduce()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`literal nil short-circuits expression (nil || "ignored")`))
		exprOk(final, null())

		orig = or(ref("some.key"), or(str("default"), ref("ignored.key")))
		final, err = orig.Reduce()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`literal "default" short-circuits expression (some.key || "default" || ignored.key)`))
		exprOk(final, or(ref("some.key"), str("default")))

		orig = or(or(ref("some.key"), str("default")), ref("ignored.key"))
		final, err = orig.Reduce()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`literal "default" short-circuits expression (some.key || "default" || ignored.key)`))
		exprOk(final, or(ref("some.key"), str("default")))
	})
})

var _ = Describe("File Operator", func() {
	var op FileOperator
	var ev *Evaluator
	var basedir string

	BeforeEach(func() {
		op = FileOperator{}
		ev = &Evaluator{
			Tree: opYAML(
				`meta:
  sample_file: assets/file_operator/sample.txt
`),
		}
		basedir, _ = os.Getwd()
	})

	It("can read a direct file", func() {
		r, err := op.Run(ev, []*Expr{
			str("assets/file_operator/test.txt"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("This is a test\n"))
	})

	It("can read a file from a reference", func() {
		r, err := op.Run(ev, []*Expr{
			ref("meta.sample_file"),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))

		content, _ := os.ReadFile("assets/file_operator/sample.txt")
		Expect(r.Value.(string)).To(Equal(string(content)))
	})

	It("can read a file relative to a specified base path", func() {
		os.Setenv("SPRUCE_FILE_BASE_PATH", filepath.Join(basedir, "assets/file_operator"))
		r, err := op.Run(ev, []*Expr{
			str("test.txt"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("This is a test\n"))
	})

	It("can read an absolute path", func() {
		if _, err := os.Stat("/etc/hosts"); err != nil {
			Skip("/etc/hosts not present on this system")
		}
		os.Setenv("SPRUCE_FILE_BASE_PATH", filepath.Join(basedir, "assets/file_operator"))
		r, err := op.Run(ev, []*Expr{
			str("/etc/hosts"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))

		content, _ := os.ReadFile("/etc/hosts")
		Expect(r.Value.(string)).To(Equal(string(content)))
	})

	It("can handle a missing file", func() {
		r, err := op.Run(ev, []*Expr{
			str("no_one_should_ever_name_a_file_that_doesnt_exist_this_name"),
		})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})
})

var _ = Describe("Grab Operator", func() {
	var op GrabOperator
	var ev *Evaluator

	BeforeEach(func() {
		op = GrabOperator{}
		ev = &Evaluator{
			Tree: opYAML(
				`key:
  subkey:
    value: found it
    other: value 2
  list1:
    - first
    - second
  list2:
    - third
    - fourth
  lonely:
    - one
`),
		}
	})

	It("can grab a single value", func() {
		r, err := op.Run(ev, []*Expr{
			ref("key.subkey.value"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("found it"))
	})

	It("can grab a single value using an environment variable in the reference", func() {
		os.Setenv("SUB_KEY", "subkey")
		r, err := op.Run(ev, []*Expr{
			ref("key.$SUB_KEY.value"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("found it"))
	})

	It("can grab a single list value", func() {
		r, err := op.Run(ev, []*Expr{
			ref("key.lonely"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))

		l, ok := r.Value.([]interface{})
		Expect(ok).To(BeTrue())

		Expect(len(l)).To(Equal(1))
		Expect(l[0]).To(Equal("one"))
	})

	It("can grab a multiple lists and flatten them", func() {
		r, err := op.Run(ev, []*Expr{
			ref("key.list1"),
			ref("key.lonely"),
			ref("key.list2"),
			ref("key.lonely.0"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))

		l, ok := r.Value.([]interface{})
		Expect(ok).To(BeTrue())

		Expect(len(l)).To(Equal(6))
		Expect(l[0]).To(Equal("first"))
		Expect(l[1]).To(Equal("second"))
		Expect(l[2]).To(Equal("one"))
		Expect(l[3]).To(Equal("third"))
		Expect(l[4]).To(Equal("fourth"))
		Expect(l[5]).To(Equal("one"))
	})

	It("can grab multiple values", func() {
		r, err := op.Run(ev, []*Expr{
			ref("key.subkey.value"),
			ref("key.subkey.other"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		v, ok := r.Value.([]interface{})
		Expect(ok).To(BeTrue())
		Expect(len(v)).To(Equal(2))
		Expect(v[0]).To(Equal("found it"))
		Expect(v[1]).To(Equal("value 2"))
	})

	It("flattens constituent arrays", func() {
		r, err := op.Run(ev, []*Expr{
			ref("key.list2"),
			ref("key.list1"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		v, ok := r.Value.([]interface{})
		Expect(ok).To(BeTrue())
		Expect(len(v)).To(Equal(4))
		Expect(v[0]).To(Equal("third"))
		Expect(v[1]).To(Equal("fourth"))
		Expect(v[2]).To(Equal("first"))
		Expect(v[3]).To(Equal("second"))
	})

	It("throws errors for missing arguments", func() {
		_, err := op.Run(ev, []*Expr{})
		Expect(err).To(HaveOccurred())
	})

	It("throws errors for dangling references", func() {
		_, err := op.Run(ev, []*Expr{
			ref("key.that.does.not.exist"),
		})
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("Environment Variable Resolution (via grab)", func() {
	var op GrabOperator
	var ev *Evaluator

	BeforeEach(func() {
		op = GrabOperator{}
		ev = &Evaluator{}
		os.Setenv("GRAB_ONE", "one")
		os.Setenv("GRAB_TWO", "two")
		os.Setenv("GRAB_NOT", "")
		os.Setenv("GRAB_BOOL", "true")
		os.Setenv("GRAB_MULTILINE", `line1

line3
line4`)
	})

	It("can grab a single environment value", func() {
		r, err := op.Run(ev, []*Expr{
			env("GRAB_ONE"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("one"))
	})

	It("tries alternates until it finds a set environment variable", func() {
		r, err := op.Run(ev, []*Expr{
			or(env("GRAB_THREE"), or(env("GRAB_TWO"), env("GRAB_ONE"))),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("two"))
	})

	It("unmarshalls variable contents", func() {
		r, err := op.Run(ev, []*Expr{
			env("GRAB_BOOL"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(bool)).To(Equal(true))
	})

	It("does not unmarshall string-only variables", func() {
		r, err := op.Run(ev, []*Expr{
			env("GRAB_MULTILINE"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal(`line1

line3
line4`))
	})

	It("throws errors for unset environment variables", func() {
		_, err := op.Run(ev, []*Expr{
			env("GRAB_NOT"),
		})
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("Concat Operator", func() {
	var op ConcatOperator
	var ev *Evaluator

	BeforeEach(func() {
		op = ConcatOperator{}
		ev = &Evaluator{
			Tree: opYAML(
				`key:
  subkey:
    value: found it
    other: value 2
  list1:
    - first
    - second
  list2:
    - third
    - fourth
douglas:
  adams: 42
math:
  PI: 3.14159
`),
		}
	})

	It("can concat a single value", func() {
		r, err := op.Run(ev, []*Expr{
			ref("key.subkey.value"),
			ref("key.list1.0"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("found itfirst"))
	})

	It("can concat a literal values", func() {
		r, err := op.Run(ev, []*Expr{
			str("a literal "),
			str("value"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("a literal value"))
	})

	It("can concat multiple values", func() {
		r, err := op.Run(ev, []*Expr{
			str("I "),
			ref("key.subkey.value"),
			str("!"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("I found it!"))
	})

	It("can concat integer literals", func() {
		r, err := op.Run(ev, []*Expr{
			str("the answer = "),
			ref("douglas.adams"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("the answer = 42"))
	})

	It("can concat float literals", func() {
		r, err := op.Run(ev, []*Expr{
			ref("math.PI"),
			str(" is PI"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("3.14159 is PI"))
	})

	It("throws errors for missing arguments", func() {
		_, err := op.Run(ev, []*Expr{})
		Expect(err).To(HaveOccurred())

		_, err = op.Run(ev, []*Expr{str("one")})
		Expect(err).To(HaveOccurred())
	})

	It("throws errors for dangling references", func() {
		_, err := op.Run(ev, []*Expr{
			ref("key.that.does.not.exist"),
			str("string"),
		})
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("static_ips Operator", func() {
	var op StaticIPOperator

	BeforeEach(func() {
		op = StaticIPOperator{}
		UsedIPs = map[string]string{}
	})

	It("can resolve valid networks inside of job contexts", func() {
		ev := &Evaluator{
			Here: cursor("jobs.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
  - name: test-network
    subnets:
      - static: [ 10.0.0.5 - 10.0.0.10 ]
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		v, ok := r.Value.([]interface{})
		Expect(ok).To(BeTrue())
		Expect(len(v)).To(Equal(3))
		Expect(v[0]).To(Equal("10.0.0.5"))
		Expect(v[1]).To(Equal("10.0.0.6"))
		Expect(v[2]).To(Equal("10.0.0.7"))
	})

	It("works with new style bosh manifests", func() {
		ev := &Evaluator{
			Here: cursor("instance_groups.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.5 - 10.0.0.10 ]
instance_groups:
- name: job1
  instances: 2
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		v, ok := r.Value.([]interface{})
		Expect(ok).To(BeTrue())
		Expect(len(v)).To(Equal(2))
		Expect(v[0]).To(Equal("10.0.0.5"))
		Expect(v[1]).To(Equal("10.0.0.6"))
	})

	It("works with multiple subnets", func() {
		ev := &Evaluator{
			Here: cursor("instance_groups.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.2 - 10.0.0.3 ]
  - static: [ 10.0.1.5 - 10.0.1.10 ]
instance_groups:
- name: job1
  instances: 4
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2), num(3)})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		v, ok := r.Value.([]interface{})
		Expect(ok).To(BeTrue())
		Expect(len(v)).To(Equal(4))
		Expect(v[0]).To(Equal("10.0.0.2"))
		Expect(v[1]).To(Equal("10.0.0.3"))
		Expect(v[2]).To(Equal("10.0.1.5"))
		Expect(v[3]).To(Equal("10.0.1.6"))
	})

	It("works with multiple subnets with an availability zone", func() {
		ev := &Evaluator{
			Here: cursor("instance_groups.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.2 - 10.0.0.3 ]
    az: z2
  - static: [ 10.0.1.5 - 10.0.1.10 ]
    az: z1
instance_groups:
- name: job1
  instances: 4
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2), num(3)})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		v, ok := r.Value.([]interface{})
		Expect(ok).To(BeTrue())
		Expect(len(v)).To(Equal(4))
		Expect(v[0]).To(Equal("10.0.0.2"))
		Expect(v[1]).To(Equal("10.0.0.3"))
		Expect(v[2]).To(Equal("10.0.1.5"))
		Expect(v[3]).To(Equal("10.0.1.6"))
	})

	It("works with instance_group availability zones", func() {
		ev := &Evaluator{
			Here: cursor("instance_groups.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.2 - 10.0.0.3 ]
    az: z1
  - static: [ 10.0.1.5 - 10.0.1.10 ]
    az: z2
instance_groups:
- name: job1
  instances: 3
  azs: [z2]
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		v, ok := r.Value.([]interface{})
		Expect(ok).To(BeTrue())
		Expect(len(v)).To(Equal(3))
		Expect(v[0]).To(Equal("10.0.1.5"))
		Expect(v[1]).To(Equal("10.0.1.6"))
		Expect(v[2]).To(Equal("10.0.1.7"))
	})

	It("works with directly specified availability zones", func() {
		ev := &Evaluator{
			Here: cursor("instance_groups.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.2 - 10.0.0.4 ]
    az: z1
  - static: [ 10.0.2.6 - 10.0.2.10 ]
    az: z2
instance_groups:
- name: job1
  instances: 6
  azs: [z1,z2]
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
		}

		r, err := op.Run(ev, []*Expr{
			str("z2:1"),
			num(0),
			str("z1:2"),
			str("z2:2"),
			num(1),
			str("z2:4"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		v, ok := r.Value.([]interface{})
		Expect(ok).To(BeTrue())
		Expect(len(v)).To(Equal(6))
		Expect(v[0]).To(Equal("10.0.2.7"))
		Expect(v[1]).To(Equal("10.0.0.2"))
		Expect(v[2]).To(Equal("10.0.0.4"))
		Expect(v[3]).To(Equal("10.0.2.8"))
		Expect(v[4]).To(Equal("10.0.0.3"))
		Expect(v[5]).To(Equal("10.0.2.10"))
	})

	It("throws an error if an unknown availability zone is used in operator", func() {
		ev := &Evaluator{
			Here: cursor("instance_groups.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.2 - 10.0.0.4 ]
    az: z1
  - static: [ 10.0.2.6 - 10.0.2.10 ]
    az: z2
instance_groups:
- name: job1
  instances: 2
  azs: [z1,z2]
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
		}

		r, err := op.Run(ev, []*Expr{
			str("z2:0"),
			str("z3:1"),
		})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("throws an error if offset for an availability zone is out of bounds", func() {
		ev := &Evaluator{
			Here: cursor("instance_groups.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.1 - 10.0.0.5 ]
    az: z1
  - static: [ 10.0.2.1 - 10.0.2.5 ]
    az: z2
instance_groups:
- name: job1
  instances: 2
  azs: [z1,z2]
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
		}

		r, err := op.Run(ev, []*Expr{
			str("z1:4"),
			str("z1:5"),
		})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("throws an error if an instance_group availability zone is not found in subnets", func() {
		ev := &Evaluator{
			Here: cursor("instance_groups.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
- name: test-net
  subnets:
  - static: [ 10.0.0.2 - 10.0.0.4 ]
    az: z1
  - static: [ 10.0.2.6 - 10.0.2.10 ]
    az: z2
instance_groups:
- name: job1
  instances: 2
  azs: [z1,z2,z3]
  networks:
  - name: test-net
    static_ips: <------------- HERE ------------
`),
		}

		r, err := op.Run(ev, []*Expr{
			str("z1:0"),
			str("z2:1"),
		})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("can resolve valid large networks inside of job contexts", func() {
		ev := &Evaluator{
			Here: cursor("jobs.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
  - name: test-network
    subnets:
      - static: [ 10.0.0.0 - 10.1.0.1 ]
jobs:
  - name: job1
    instances: 7
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
		}

		r, err := op.Run(ev, []*Expr{
			num(0),
			num(255),   // 2^8 - 1
			num(256),   // 2^8
			num(257),   // 2^8 + 1
			num(65535), // 2^16 - 1
			num(65536), // 2^16
			num(65537), // 2^16 + 1
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		v, ok := r.Value.([]interface{})
		Expect(ok).To(BeTrue())
		Expect(len(v)).To(Equal(7))
		Expect(v[0]).To(Equal("10.0.0.0"))
		Expect(v[1]).To(Equal("10.0.0.255"))
		Expect(v[2]).To(Equal("10.0.1.0")) //  3rd octet rollover
		Expect(v[3]).To(Equal("10.0.1.1"))

		Expect(v[4]).To(Equal("10.0.255.255"))
		Expect(v[5]).To(Equal("10.1.0.0")) //  2nd octet rollover
		Expect(v[6]).To(Equal("10.1.0.1"))
	})

	It("throws an error if no job name is specified", func() {
		ev := &Evaluator{
			Here: cursor("jobs.0.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
  - name: test-network
    subnets:
      - static: [ 10.0.0.5 - 10.0.0.10 ]
jobs:
  - instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("throws an error if no job instances specified", func() {
		ev := &Evaluator{
			Here: cursor("jobs.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
  - name: test-network
    subnets:
      - static: [ 10.0.0.5 - 10.0.0.10 ]
jobs:
  - name: job1
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("throws an error if job instances is not a number", func() {
		ev := &Evaluator{
			Here: cursor("jobs.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
  - name: test-network
    subnets:
      - static: [ 10.0.0.5 - 10.0.0.10 ]
jobs:
  - name: job1
    instances: PI
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("throws an error if job has no network name", func() {
		ev := &Evaluator{
			Here: cursor("jobs.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
  - name: test-network
    subnets:
      - static: [ 10.0.0.5 - 10.0.0.10 ]
jobs:
  - name: job1
    instances: 3
    networks:
      - static_ips: <---------- HERE -----------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("throws an error if network has no subnets key", func() {
		ev := &Evaluator{
			Here: cursor("jobs.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
  - name: test-network
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("throws an error if network has no subnets", func() {
		ev := &Evaluator{
			Here: cursor("jobs.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
  - name: test-network
    subnets: []
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("throws an error if network has no static ranges", func() {
		ev := &Evaluator{
			Here: cursor("jobs.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
  - name: test-network
    subnets:
     - {}
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("throws an error if network has malformed static range array(s)", func() {
		ev := &Evaluator{
			Here: cursor("jobs.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
  - name: test-network
    subnets:
    - static: [ 10.0.0.1, 10.0.0.254 ]
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("throws an error if network static range has malformed IP addresses", func() {
		ev := &Evaluator{
			Here: cursor("jobs.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
  - name: test-network
    subnets:
    - static: 10.0.0.0.0.0.0.1 - geoff
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("throws an error if the static address pool is too small", func() {
		ev := &Evaluator{
			Here: cursor("jobs.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
  - name: test-network
    subnets:
    - static: 172.16.31.10 - 172.16.31.11
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("throws an error if the address pool ends before it starts", func() {
		ev := &Evaluator{
			Here: cursor("jobs.job1.networks.0.static_ips"),
			Tree: opYAML(
				`networks:
  - name: test-network
    subnets:
    - static: [ 10.8.0.1 - 10.0.0.255 ]
jobs:
  - name: job1
    instances: 3
    networks:
      - name: test-network
        static_ips: <---------- HERE -----------------
`),
		}

		r, err := op.Run(ev, []*Expr{num(0), num(1), num(2)})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("ends before it starts"))
		Expect(r).To(BeNil())
	})
})

var _ = Describe("inject Operator", func() {
	var op InjectOperator
	var ev *Evaluator

	BeforeEach(func() {
		op = InjectOperator{}
		ev = &Evaluator{
			Tree: opYAML(
				`key:
  subkey:
    value: found it
    other: value 2
  subkey2:
    value: overridden
    third: trois
  list1:
    - first
    - second
  list2:
    - third
    - fourth
`),
		}
	})

	It("can inject a single sub-map", func() {
		r, err := op.Run(ev, []*Expr{
			ref("key.subkey"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Inject))
		v, ok := r.Value.(map[interface{}]interface{})
		Expect(ok).To(BeTrue())
		Expect(v["value"]).To(Equal("found it"))
		Expect(v["other"]).To(Equal("value 2"))
	})

	It("can inject multiple sub-maps", func() {
		r, err := op.Run(ev, []*Expr{
			ref("key.subkey"),
			ref("key.subkey2"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Inject))
		v, ok := r.Value.(map[interface{}]interface{})
		Expect(ok).To(BeTrue())
		Expect(len(v)).To(Equal(3))
		Expect(v["value"]).To(Equal("overridden"))
		Expect(v["other"]).To(Equal("value 2"))
		Expect(v["third"]).To(Equal("trois"))
	})

	It("handles non-existent references", func() {
		_, err := op.Run(ev, []*Expr{
			ref("key.subkey"),
			ref("key.subkey2"),
			ref("key.subkey2.ENOENT"),
		})
		Expect(err).To(HaveOccurred())
	})

	It("throws an error when trying to inject a scalar", func() {
		_, err := op.Run(ev, []*Expr{
			ref("key.subkey.value"),
		})
		Expect(err).To(HaveOccurred())
	})

	It("throws an error when trying to inject a list", func() {
		_, err := op.Run(ev, []*Expr{
			ref("key.list1"),
		})
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("param Operator", func() {
	var op ParamOperator
	var ev *Evaluator

	BeforeEach(func() {
		op = ParamOperator{}
		ev = &Evaluator{}
	})

	It("always causes an error", func() {
		r, err := op.Run(ev, []*Expr{
			str("this is the error"),
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("this is the error"))
		Expect(r).To(BeNil())
	})
})

var _ = Describe("Join Operator", func() {
	var op JoinOperator
	var ev *Evaluator

	BeforeEach(func() {
		op = JoinOperator{}
		ev = &Evaluator{
			Tree: opYAML(
				`---
meta:
  authorities:
  - password.write
  - clients.write
  - clients.read
  - scim.write
  - scim.read
  - uaa.admin
  - clients.secret

  secondlist:
  - admin.write
  - admin.read

  emptylist: []

  anotherkey:
  - entry1
  - somekey: value
  - entry2

  somestanza:
    foo: bar
    wom: bat
`),
		}
	})

	It("can join a simple list", func() {
		r, err := op.Run(ev, []*Expr{
			str(","),
			ref("meta.authorities"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("password.write,clients.write,clients.read,scim.write,scim.read,uaa.admin,clients.secret"))
	})

	It("can join multiple lists", func() {
		r, err := op.Run(ev, []*Expr{
			str(","),
			ref("meta.authorities"),
			ref("meta.secondlist"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("password.write,clients.write,clients.read,scim.write,scim.read,uaa.admin,clients.secret,admin.write,admin.read"))
	})

	It("can join an empty list", func() {
		r, err := op.Run(ev, []*Expr{
			str(","),
			ref("meta.emptylist"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal(""))
	})

	It("can join string literals", func() {
		r, err := op.Run(ev, []*Expr{
			str(","),
			str("password.write"),
			str("clients.write"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("password.write,clients.write"))
	})

	It("can join integer literals", func() {
		r, err := op.Run(ev, []*Expr{
			str(":"),
			num(4), num(8), num(15),
			num(16), num(23), num(42),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("4:8:15:16:23:42"))
	})

	It("can join referenced string entry", func() {
		r, err := op.Run(ev, []*Expr{
			str(","),
			ref("meta.somestanza.foo"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("bar"))
	})

	It("can join referenced string entries", func() {
		r, err := op.Run(ev, []*Expr{
			str(","),
			ref("meta.somestanza.foo"),
			ref("meta.somestanza.wom"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("bar,bat"))
	})

	It("can join multiple referenced entries", func() {
		r, err := op.Run(ev, []*Expr{
			str(","),
			ref("meta.authorities"),
			ref("meta.somestanza.foo"),
			ref("meta.somestanza.wom"),
			str("ending"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("password.write,clients.write,clients.read,scim.write,scim.read,uaa.admin,clients.secret,bar,bat,ending"))
	})

	It("throws an error when there are no arguments", func() {
		r, err := op.Run(ev, []*Expr{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no arguments specified"))
		Expect(r).To(BeNil())
	})

	It("throws an error when there are too few arguments", func() {
		r, err := op.Run(ev, []*Expr{
			str(","),
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("too few arguments supplied"))
		Expect(r).To(BeNil())
	})

	It("throws an error when separator argument is not a literal", func() {
		r, err := op.Run(ev, []*Expr{
			ref("meta.emptylist"),
			ref("meta.authorities"),
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("join operator only accepts literal argument for the separator"))
		Expect(r).To(BeNil())
	})

	It("throws an error when referenced entry is not a list or literal", func() {
		r, err := op.Run(ev, []*Expr{
			str(","),
			ref("meta.somestanza"),
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("referenced entry is not a list or string"))
		Expect(r).To(BeNil())
	})

	It("throws an error when referenced list contains non-string entries", func() {
		r, err := op.Run(ev, []*Expr{
			str(","),
			ref("meta.anotherkey"),
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("is not compatible for"))
		Expect(r).To(BeNil())
	})

	It("throws an error when there are unresolvable references", func() {
		r, err := op.Run(ev, []*Expr{
			str(","),
			ref("meta.non-existent"),
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unable to resolve"))
		Expect(r).To(BeNil())
	})

	Describe("calculates dependencies correctly", func() {
		shouldHaveDeps := func(actual interface{}, expected ...interface{}) string {
			deps := actual.([]*tree.Cursor)
			paths := []string{}
			for _, path := range expected {
				normalizedPath, err := tree.ParseCursor(path.(string))
				if err != nil {
					panic(fmt.Sprintf("improper path %s passed to test", path.(string)))
				}
				paths = append(paths, normalizedPath.String())
			}
			actualPaths := []string{}
			for _, dep := range deps {
				actualPaths = append(actualPaths, dep.String())
			}
			sort.Strings(actualPaths)
			sort.Strings(paths)
			match := reflect.DeepEqual(actualPaths, paths)
			if !match {
				return fmt.Sprintf("actual: %+v\n expected: %+v", actualPaths, paths)
			}
			return ""
		}

		It("with a single list", func() {
			deps := op.Dependencies(ev, []*Expr{
				str(" "),
				ref("meta.secondlist"),
			},
				nil,
				nil)
			Expect(deps).To(WithTransform(func(d []*tree.Cursor) interface{} { return d }, Satisfy(func(v interface{}) bool {
				return shouldHaveDeps(v, "meta.secondlist.[0]", "meta.secondlist.[1]") == ""
			})))
		})

		It("with multiple lists", func() {
			deps := op.Dependencies(ev, []*Expr{
				str(" "),
				ref("meta.authorities"),
				ref("meta.secondlist"),
			},
				nil,
				nil)
			Expect(deps).To(WithTransform(func(d []*tree.Cursor) interface{} { return d }, Satisfy(func(v interface{}) bool {
				return shouldHaveDeps(v, "meta.authorities.[0]", "meta.authorities.[1]",
					"meta.authorities.[2]", "meta.authorities.[3]", "meta.authorities.[4]",
					"meta.authorities.[5]", "meta.authorities.[6]",
					"meta.secondlist.[0]", "meta.secondlist.[1]") == ""
			})))
		})

		It("with a reference string", func() {
			deps := op.Dependencies(ev, []*Expr{
				str(" "),
				ref("meta.somestanza.foo"),
			},
				nil,
				nil)
			Expect(deps).To(WithTransform(func(d []*tree.Cursor) interface{} { return d }, Satisfy(func(v interface{}) bool {
				return shouldHaveDeps(v, "meta.somestanza.foo") == ""
			})))
		})

		It("with multiple reference strings", func() {
			deps := op.Dependencies(ev, []*Expr{
				str(" "),
				ref("meta.somestanza.foo"),
				ref("meta.somestanza.wom"),
			},
				nil,
				nil)
			Expect(deps).To(WithTransform(func(d []*tree.Cursor) interface{} { return d }, Satisfy(func(v interface{}) bool {
				return shouldHaveDeps(v, "meta.somestanza.foo", "meta.somestanza.wom") == ""
			})))
		})

		It("with a reference string and a list", func() {
			deps := op.Dependencies(ev, []*Expr{
				str(" "),
				ref("meta.somestanza.foo"),
				ref("meta.secondlist"),
			},
				nil,
				nil)
			Expect(deps).To(WithTransform(func(d []*tree.Cursor) interface{} { return d }, Satisfy(func(v interface{}) bool {
				return shouldHaveDeps(v, "meta.somestanza.foo", "meta.secondlist.[0]",
					"meta.secondlist.[1]") == ""
			})))
		})

		It("with a literal string", func() {
			deps := op.Dependencies(ev, []*Expr{
				str(" "),
				str("literally literal"),
			},
				nil,
				nil)
			Expect(deps).To(WithTransform(func(d []*tree.Cursor) interface{} { return d }, Satisfy(func(v interface{}) bool {
				return shouldHaveDeps(v) == ""
			})))
		})

		It("with a literal string and a reference string", func() {
			deps := op.Dependencies(ev, []*Expr{
				str(" "),
				str("beep"),
				ref("meta.somestanza.foo"),
			},
				nil,
				nil)
			Expect(deps).To(WithTransform(func(d []*tree.Cursor) interface{} { return d }, Satisfy(func(v interface{}) bool {
				return shouldHaveDeps(v, "meta.somestanza.foo") == ""
			})))
		})
	})
})

var _ = Describe("empty operator", func() {
	var op EmptyOperator
	var ev *Evaluator

	BeforeEach(func() {
		op = EmptyOperator{}
		ev = &Evaluator{
			Tree: opYAML(
				`---
meta:
  authorities: meep
`),
		}
	})

	It("can replace with a hash", func() {
		r, err := op.Run(ev, []*Expr{
			ref("hash"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		val, isHash := r.Value.(map[string]interface{})
		Expect(isHash).To(BeTrue())
		Expect(val).To(Equal(map[string]interface{}{}))
	})

	It("can replace with an array", func() {
		r, err := op.Run(ev, []*Expr{
			ref("array"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		val, isArray := r.Value.([]interface{})
		Expect(isArray).To(BeTrue())
		Expect(val).To(Equal([]interface{}{}))
	})

	It("can replace with an empty string", func() {
		r, err := op.Run(ev, []*Expr{
			ref("string"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		val, isString := r.Value.(string)
		Expect(isString).To(BeTrue())
		Expect(val).To(Equal(""))
	})

	It("throws an error for unrecognized types", func() {
		r, err := op.Run(ev, []*Expr{
			ref("void"),
		})
		Expect(r).To(BeNil())
		Expect(err).To(HaveOccurred())
	})

	It("works with string literals", func() {
		r, err := op.Run(ev, []*Expr{
			str("hash"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		val, isHash := r.Value.(map[string]interface{})
		Expect(isHash).To(BeTrue())
		Expect(val).To(Equal(map[string]interface{}{}))
	})

	It("throws an error with no args", func() {
		r, err := op.Run(ev, []*Expr{})
		Expect(r).To(BeNil())
		Expect(err).To(HaveOccurred())
	})

	It("throws an error with too many args", func() {
		r, err := op.Run(ev, []*Expr{
			ref("hash"),
			ref("array"),
		})
		Expect(r).To(BeNil())
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("ips Operator", func() {
	var op IpsOperator
	var ev *Evaluator

	BeforeEach(func() {
		op = IpsOperator{}
		ev = &Evaluator{
			Tree: opYAML(
				`meta:
  base_network: 1.2.3.4/24
  base_ip: 1.2.3.4
  index: 20
  negative_index: -20
  count: 2
`),
		}
	})

	It("can build a single IP based on refs (CIDR)", func() {
		r, err := op.Run(ev, []*Expr{
			ref("meta.base_network"),
			ref("meta.index"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("1.2.3.20"))
	})

	It("can build a single IP based on refs (IP)", func() {
		r, err := op.Run(ev, []*Expr{
			ref("meta.base_ip"),
			ref("meta.index"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("1.2.3.24"))
	})

	It("can build a single IP based on literals", func() {
		r, err := op.Run(ev, []*Expr{
			str("1.2.3.4/24"),
			num(20),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("1.2.3.20"))
	})

	It("can build a list of IP's based on references (CIDR)", func() {
		r, err := op.Run(ev, []*Expr{
			ref("meta.base_network"),
			ref("meta.index"),
			ref("meta.count"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.([]interface{})[0].(string)).To(Equal("1.2.3.20"))
		Expect(r.Value.([]interface{})[1].(string)).To(Equal("1.2.3.21"))
		Expect(len(r.Value.([]interface{}))).To(Equal(2))
	})

	It("can build a list of IP's based on literals", func() {
		r, err := op.Run(ev, []*Expr{
			str("1.2.3.4/24"),
			num(20),
			num(2),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.([]interface{})[0].(string)).To(Equal("1.2.3.20"))
		Expect(r.Value.([]interface{})[1].(string)).To(Equal("1.2.3.21"))
		Expect(len(r.Value.([]interface{}))).To(Equal(2))
	})

	It("can build a list of IP's based on references (IP)", func() {
		r, err := op.Run(ev, []*Expr{
			ref("meta.base_ip"),
			ref("meta.index"),
			ref("meta.count"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.([]interface{})[0].(string)).To(Equal("1.2.3.24"))
		Expect(r.Value.([]interface{})[1].(string)).To(Equal("1.2.3.25"))
		Expect(len(r.Value.([]interface{}))).To(Equal(2))
	})

	It("can build a list of IP's using negative index (IP)", func() {
		r, err := op.Run(ev, []*Expr{
			ref("meta.base_ip"),
			ref("meta.negative_index"),
			ref("meta.count"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.([]interface{})[0].(string)).To(Equal("1.2.2.240"))
		Expect(r.Value.([]interface{})[1].(string)).To(Equal("1.2.2.241"))
		Expect(len(r.Value.([]interface{}))).To(Equal(2))
	})

	It("can build a list of IP's using negative index (CIDR)", func() {
		r, err := op.Run(ev, []*Expr{
			ref("meta.base_network"),
			ref("meta.negative_index"),
			ref("meta.count"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.([]interface{})[0].(string)).To(Equal("1.2.3.236"))
		Expect(r.Value.([]interface{})[1].(string)).To(Equal("1.2.3.237"))
		Expect(len(r.Value.([]interface{}))).To(Equal(2))
	})

	It("bails out if index is outside CIDR size", func() {
		r, err := op.Run(ev, []*Expr{
			str("192.168.1.16/29"),
			num(100),
		})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("start index 100 exceeds size of subnet 192.168.1.16/29"))
		Expect(r).To(BeNil())
	})

	It("bails out if count would go outside CIDR size", func() {
		r, err := op.Run(ev, []*Expr{
			str("192.168.1.16/29"),
			num(-1),
			num(3),
		})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("start index 7 and count 3 would exceed size of subnet 192.168.1.16/29"))
		Expect(r).To(BeNil())
	})
})

var _ = Describe("Base64 Operator", func() {
	var op Base64Operator
	var ev *Evaluator

	BeforeEach(func() {
		op = Base64Operator{}
		ev = &Evaluator{
			Tree: opYAML(
				`meta:
  sample: "Sample Text To Base64 Encode From Reference"
`),
		}
	})

	It("can encode a string literal", func() {
		r, err := op.Run(ev, []*Expr{
			str("Sample Text To Base64 Encode"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("U2FtcGxlIFRleHQgVG8gQmFzZTY0IEVuY29kZQ=="))
	})

	It("can encode from a reference", func() {
		r, err := op.Run(ev, []*Expr{
			ref("meta.sample"),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("U2FtcGxlIFRleHQgVG8gQmFzZTY0IEVuY29kZSBGcm9tIFJlZmVyZW5jZQ=="))
	})

	It("can handle non string scalar input", func() {
		r, err := op.Run(ev, []*Expr{
			str("one"), num(1),
		})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("can handle non string scalar input (i.e numbers)", func() {
		r, err := op.Run(ev, []*Expr{
			num(1),
		})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})
})

var _ = Describe("Base64Decode Operator", func() {
	var op Base64DecodeOperator
	var ev *Evaluator

	BeforeEach(func() {
		op = Base64DecodeOperator{}
		ev = &Evaluator{
			Tree: opYAML(
				`meta:
  sample: "U2FtcGxlIFRleHQgVG8gQmFzZTY0IEVuY29kZSBGcm9tIFJlZmVyZW5jZQ=="
`),
		}
	})

	It("can decode from a string literal", func() {
		r, err := op.Run(ev, []*Expr{
			str("U2FtcGxlIFRleHQgVG8gQmFzZTY0IEVuY29kZQ=="),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("Sample Text To Base64 Encode"))
	})

	It("can decode from a reference", func() {
		r, err := op.Run(ev, []*Expr{
			ref("meta.sample"),
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("Sample Text To Base64 Encode From Reference"))
	})
})

var _ = Describe("awsparam/awssecret operator", func() {
	var op AwsOperator
	var ev *Evaluator
	var fakeSSM *fakes.FakeSSMAPI
	var fakeSecretsManager *fakes.FakeSecretsManagerAPI

	BeforeEach(func() {
		op = AwsOperator{variant: "awsparam"}
		ev = &Evaluator{
			Tree: opYAML(`{ "testval": "test", "testmap": {}, "testarr": [] }`),
			Here: &tree.Cursor{},
		}
		fakeSSM = new(fakes.FakeSSMAPI)
		fakeSecretsManager = new(fakes.FakeSecretsManagerAPI)
		parameterstoreClient = fakeSSM
		secretsManagerClient = fakeSecretsManager
	})

	Describe("in shared logic", func() {
		It("should return error if no key given", func() {
			_, err := op.Run(ev, []*Expr{})
			Expect(err.Error()).To(ContainSubstring("awsparam operator requires at least one argument"))
		})

		It("should concatenate args", func() {
			var ssmKey string
			fakeSSM.GetParameterStub = func(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
				ssmKey = aws.StringValue(in.Name)
				return &ssm.GetParameterOutput{
					Parameter: &ssm.Parameter{
						Value: aws.String(""),
					},
				}, nil
			}
			_, err := op.Run(ev, []*Expr{num(1), num(2), num(3)})
			Expect(err).NotTo(HaveOccurred())
			Expect(ssmKey).To(Equal("123"))
		})

		It("should resolve references", func() {
			var ssmKey string
			fakeSSM.GetParameterStub = func(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
				ssmKey = aws.StringValue(in.Name)
				return &ssm.GetParameterOutput{
					Parameter: &ssm.Parameter{
						Value: aws.String(""),
					},
				}, nil
			}
			_, err := op.Run(ev, []*Expr{num(1), num(2), ref("testval")})
			Expect(err).NotTo(HaveOccurred())
			Expect(ssmKey).To(Equal("12test"))
		})

		It("should not allow references to maps", func() {
			_, err := op.Run(ev, []*Expr{num(1), num(2), ref("testmap")})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("$.testmap is a map; only scalars are supported here"))
		})

		It("should not allow references to arrays", func() {
			_, err := op.Run(ev, []*Expr{num(1), num(2), ref("testarr")})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("$.testarr is a list; only scalars are supported here"))
		})

		It("without key", func() {
			fakeSSM.GetParameterStub = func(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
				return &ssm.GetParameterOutput{
					Parameter: &ssm.Parameter{
						Value: aws.String("testx"),
					},
				}, nil
			}
			r, err := op.Run(ev, []*Expr{str("val1")})
			Expect(err).NotTo(HaveOccurred())
			Expect(r.Type).To(Equal(Replace))
			Expect(r.Value.(string)).To(Equal("testx"))
		})

		Describe("with key", func() {
			It("should parse subkey and extract if provided", func() {
				fakeSSM.GetParameterStub = func(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
					return &ssm.GetParameterOutput{
						Parameter: &ssm.Parameter{
							Value: aws.String(`{ "key": "val" }`),
						},
					}, nil
				}
				r, err := op.Run(ev, []*Expr{str("val2?key=key")})
				Expect(err).NotTo(HaveOccurred())
				Expect(r.Type).To(Equal(Replace))
				Expect(r.Value.(string)).To(Equal("val"))
			})

			It("should error if document not valid yaml / json", func() {
				fakeSSM.GetParameterStub = func(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
					return &ssm.GetParameterOutput{
						Parameter: &ssm.Parameter{
							Value: aws.String(`key: {`),
						},
					}, nil
				}
				_, err := op.Run(ev, []*Expr{str("val3?key=key")})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("$.val3 error extracting key: yaml: line 1: did not find expected node content"))
			})

			It("should error if subkey invalid", func() {
				fakeSSM.GetParameterStub = func(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
					return &ssm.GetParameterOutput{
						Parameter: &ssm.Parameter{
							Value: aws.String(`key: {}`),
						},
					}, nil
				}
				_, err := op.Run(ev, []*Expr{str("val4?key=noexist")})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("$.val4 invalid key 'noexist'"))
			})
		})

		It("should not call AWS API if SkipAws true", func() {
			SkipAws = true
			defer func() { SkipAws = false }()
			count := 0
			fakeSSM.GetParameterStub = func(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
				count++
				return &ssm.GetParameterOutput{
					Parameter: &ssm.Parameter{
						Value: aws.String(""),
					},
				}, nil
			}
			_, err := op.Run(ev, []*Expr{str("skipaws")})
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(0))
		})
	})

	Describe("awsparam", func() {
		It("should cache lookups", func() {
			count := 0
			fakeSSM.GetParameterStub = func(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
				count++
				return &ssm.GetParameterOutput{
					Parameter: &ssm.Parameter{
						Value: aws.String(""),
					},
				}, nil
			}
			_, err := op.Run(ev, []*Expr{str("val5")})
			Expect(err).NotTo(HaveOccurred())
			_, err = op.Run(ev, []*Expr{str("val5")})
			Expect(err).NotTo(HaveOccurred())

			Expect(count).To(Equal(1))
		})
	})

	Describe("awssecret", func() {
		BeforeEach(func() {
			op = AwsOperator{variant: "awssecret"}
		})

		It("should cache lookups", func() {
			count := 0
			fakeSecretsManager.GetSecretValueStub = func(in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
				count++
				return &secretsmanager.GetSecretValueOutput{
					SecretString: aws.String(""),
				}, nil
			}
			_, err := op.Run(ev, []*Expr{str("val5")})
			Expect(err).NotTo(HaveOccurred())
			_, err = op.Run(ev, []*Expr{str("val5")})
			Expect(err).NotTo(HaveOccurred())

			Expect(count).To(Equal(1))
		})

		It("should use stage if provided", func() {
			stage := ""
			fakeSecretsManager.GetSecretValueStub = func(in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
				stage = aws.StringValue(in.VersionStage)
				return &secretsmanager.GetSecretValueOutput{
					SecretString: aws.String(""),
				}, nil
			}
			_, err := op.Run(ev, []*Expr{str("val6?stage=test")})
			Expect(err).NotTo(HaveOccurred())

			Expect(stage).To(Equal("test"))
		})

		It("should use version if provided", func() {
			version := ""
			fakeSecretsManager.GetSecretValueStub = func(in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
				version = aws.StringValue(in.VersionId)
				return &secretsmanager.GetSecretValueOutput{
					SecretString: aws.String(""),
				}, nil
			}
			_, err := op.Run(ev, []*Expr{str("val7?version=test")})
			Expect(err).NotTo(HaveOccurred())

			Expect(version).To(Equal("test"))
		})
	})
})

var _ = Describe("Stringify Operator", func() {
	var op StringifyOperator
	var ev *Evaluator

	BeforeEach(func() {
		op = StringifyOperator{}
		ev = &Evaluator{
			Tree: opYAML(`meta:
  map:
    bar: foo
    foo: bar
  list:
  - first
  - second
  scalars:
    bool: true
    number: 42
    string: foobar
`),
		}
	})

	It("cannot use operator with more than one reference", func() {
		r, err := op.Run(ev, []*Expr{
			ref("meta.map"),
			ref("list.0"),
		})
		Expect(err).To(HaveOccurred())
		Expect(r).To(BeNil())
	})

	It("can stringify map", func() {
		r, err := op.Run(ev, []*Expr{
			ref("meta.map"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal(`bar: foo
foo: bar
`))
	})

	It("can stringify list", func() {
		r, err := op.Run(ev, []*Expr{
			ref("meta.list"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal(`- first
- second
`))
	})

	It("can stringify scalars", func() {
		r, err := op.Run(ev, []*Expr{
			ref("meta.scalars"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal(`bool: true
number: 42
string: foobar
`))
	})

	It("retain string literal", func() {
		r, err := op.Run(ev, []*Expr{
			str("foo"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("foo"))
	})

	It("retain null literal", func() {
		r, err := op.Run(ev, []*Expr{
			null(),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())

		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value).To(BeNil())
	})

	It("throws errors for dangling references", func() {
		_, err := op.Run(ev, []*Expr{
			ref("key.that.does.not.exist"),
		})
		Expect(err).To(HaveOccurred())
	})

	It("throws errors for missing arguments", func() {
		_, err := op.Run(ev, []*Expr{})
		Expect(err).To(HaveOccurred())
	})
})
