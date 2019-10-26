package spruce

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"

	"github.com/geofffranks/simpleyaml"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/starkandwayne/goutils/tree"
)

type mockedAws struct {
	ssmiface.SSMAPI
	secretsmanageriface.SecretsManagerAPI

	MockGetSecretValue func(*secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
	MockGetParameter   func(*ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
}

func (m *mockedAws) GetSecretValue(input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	return m.MockGetSecretValue(input)
}

func (m *mockedAws) GetParameter(input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	return m.MockGetParameter(input)
}

func TestOperators(t *testing.T) {
	cursor := func(s string) *tree.Cursor {
		c, err := tree.ParseCursor(s)
		So(err, ShouldBeNil)
		return c
	}

	YAML := func(s string) map[interface{}]interface{} {
		y, err := simpleyaml.NewYaml([]byte(s))
		So(err, ShouldBeNil)

		data, err := y.Map()
		So(err, ShouldBeNil)

		return data
	}

	ref := func(s string) *Expr {
		return &Expr{Type: Reference, Reference: cursor(s)}
	}
	str := func(s string) *Expr {
		return &Expr{Type: Literal, Literal: s}
	}
	num := func(v int64) *Expr {
		return &Expr{Type: Literal, Literal: v}
	}
	null := func() *Expr {
		return &Expr{Type: Literal, Literal: nil}
	}
	env := func(s string) *Expr {
		return &Expr{Type: EnvVar, Name: s}
	}
	or := func(l *Expr, r *Expr) *Expr {
		return &Expr{Type: LogicalOr, Left: l, Right: r}
	}

	var exprOk func(*Expr, *Expr)
	exprOk = func(got *Expr, want *Expr) {
		So(got, ShouldNotBeNil)
		So(want, ShouldNotBeNil)

		So(got.Type, ShouldEqual, want.Type)
		switch want.Type {
		case Literal:
			So(got.Literal, ShouldEqual, want.Literal)

		case Reference:
			So(got.Reference.String(), ShouldEqual, want.Reference.String())

		case LogicalOr:
			exprOk(got.Left, want.Left)
			exprOk(got.Right, want.Right)
		}
	}

	Convey("Parser", t, func() {
		Convey("parses op calls in their entirety", func() {
			phase := EvalPhase

			opOk := func(code string, name string, args ...*Expr) {
				op, err := ParseOpcall(phase, code)
				So(err, ShouldBeNil)
				So(op, ShouldNotBeNil)

				_, ok := op.op.(NullOperator)
				So(ok, ShouldBeTrue)
				So(op.op.(NullOperator).Missing, ShouldEqual, name)

				So(len(op.args), ShouldEqual, len(args))
				for i, expect := range args {
					exprOk(op.args[i], expect)
				}
			}

			opErr := func(code string, msg string) {
				_, err := ParseOpcall(phase, code)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, msg)
			}

			opIgnore := func(code string) {
				op, err := ParseOpcall(phase, code)
				So(op, ShouldBeNil)
				So(err, ShouldBeNil)
			}

			Convey("handles opcodes with and without arguments", func() {
				opOk(`(( null 42 ))`, "null", num(42))
				opOk(`(( null 1 2 3 4 ))`, "null", num(1), num(2), num(3), num(4))
			})

			Convey("ignores optional whitespace", func() {
				args := []*Expr{num(1), num(2), num(3)}
				opOk(`((null 1 2 3))`, "null", args...)
				opOk(`((null 1	2	3))`, "null", args...)
				opOk(`((null 1	2	3	))`, "null", args...)
				opOk(`((null 1 	 2 	 3 	 ))`, "null", args...)
			})

			Convey("allows use of commas to separate arguments", func() {
				args := []*Expr{num(1), num(2), num(3)}
				opOk(`((null 1, 2, 3))`, "null", args...)
				opOk(`((null 1,	2,	3))`, "null", args...)
				opOk(`((null 1,	2,	3,	))`, "null", args...)
				opOk(`((null 1 ,	 2 ,	 3 ,	 ))`, "null", args...)
			})

			Convey("allows use of parentheses around arguments", func() {
				args := []*Expr{num(1), num(2), num(3)}
				opOk(`((null(1,2,3)))`, "null", args...)
				opOk(`((null(1, 2, 3) ))`, "null", args...)
				opOk(`((null( 1,	2,	3)))`, "null", args...)
				opOk(`((null (1,	2,	3)	))`, "null", args...)
				opOk(`((null (1 ,	 2 ,	 3)	 ))`, "null", args...)
			})

			Convey("handles string literal arguments", func() {
				opOk(`(( null "string" ))`, "null", str("string"))
				opOk(`(( null "string with whitespace" ))`, "null", str("string with whitespace"))
				opOk(`(( null "a \"quoted\" string" ))`, "null", str(`a "quoted" string`))
				opOk(`(( null "\\escaped" ))`, "null", str(`\escaped`))
			})

			Convey("handles reference (cursor) arguments", func() {
				opOk(`(( null x.y.z ))`, "null", ref("x.y.z"))
				opOk(`(( null x.[0].z ))`, "null", ref("x.0.z"))
				opOk(`(( null x[0].z ))`, "null", ref("x.0.z"))
				opOk(`(( null x[0]z ))`, "null", ref("x.0.z"))
			})

			Convey("handles mixed collections of argument types", func() {
				opOk(`(( xyzzy "string" x.y.z 42  ))`, "xyzzy", str("string"), ref("x.y.z"), num(42))
				opOk(`(( xyzzy("string" x.y.z 42) ))`, "xyzzy", str("string"), ref("x.y.z"), num(42))
			})

			Convey("handles expression-based operands", func() {
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

			Convey("handles environment variables as operands", func() {
				os.Setenv("SPRUCE_FOO", "first test")
				os.Setenv("_SPRUCE", "_sprucify!")
				os.Setenv("ENOENT", "")
				os.Setenv("http_proxy", "no://thank/you")
				os.Setenv("variable.with.dots", "dots are ok")

				opOk(`(( null $SPRUCE_FOO ))`, "null", env("SPRUCE"))
				opOk(`(( null $_SPRUCE ))`, "null", env("_SPRUCE"))
				opOk(`(( null $ENOENT || $SPRUCE_FOO ))`, "null",
					or(env("ENOENT"), env("SPRUCE_FOO")))
				opOk(`(( null $http_proxy))`, "null", env("http_proxy"))
				opOk(`(( null $variable.with.dots ))`, "null", env("variable.with.dots"))
			})

			Convey("throws errors for malformed expression", func() {
				opErr(`(( null meta.key ||, nil ))`,
					`syntax error near: meta.key ||, nil`)

				opErr(`(( null || ))`,
					`syntax error near: ||`)

				opErr(`(( null || meta.key ))`,
					`syntax error near: || meta.key`)

				opErr(`(( null meta.key || || ))`,
					`syntax error near: meta.key || ||`)
			})

			Convey("ignores spiff-like bang-notation", func() {
				opIgnore(`((!credhub))`)
			})

			Convey("ignores BOSH varnames that aren't null-arity operators", func() {
				opIgnore(`((var-name))`)
			})
		})
	})

	Convey("Expression Engine", t, func() {
		var e *Expr
		var tree map[interface{}]interface{}

		evaluate := func(e *Expr, tree map[interface{}]interface{}) interface{} {
			v, err := e.Evaluate(tree)
			So(err, ShouldBeNil)
			return v
		}

		Convey("Literals evaluate to themselves", func() {
			e = &Expr{Type: Literal, Literal: "value"}
			So(evaluate(e, tree), ShouldEqual, "value")

			e = &Expr{Type: Literal, Literal: ""}
			So(evaluate(e, tree), ShouldEqual, "")

			e = &Expr{Type: Literal, Literal: nil}
			So(evaluate(e, tree), ShouldEqual, nil)
		})

		Convey("References evaluate to the referenced part of the YAML tree", func() {
			tree = YAML(`---
meta:
  foo: FOO
  bar: BAR
`)

			e = &Expr{Type: Reference, Reference: cursor("meta.foo")}
			So(evaluate(e, tree), ShouldEqual, "FOO")

			e = &Expr{Type: Reference, Reference: cursor("meta.bar")}
			So(evaluate(e, tree), ShouldEqual, "BAR")
		})

		Convey("|| operator evaluates to the first found value", func() {
			tree = YAML(`---
meta:
  foo: FOO
  bar: BAR
`)

			So(evaluate(or(str("first"), str("second")), tree), ShouldEqual, "first")
			So(evaluate(or(ref("meta.foo"), str("second")), tree), ShouldEqual, "FOO")
			So(evaluate(or(ref("meta.ENOENT"), ref("meta.foo")), tree), ShouldEqual, "FOO")
		})

		Convey("|| operator treats nil as a found value", func() {
			tree = YAML(`---
meta:
  foo: FOO
  bar: BAR
`)

			So(evaluate(or(null(), str("second")), tree), ShouldBeNil)
			So(evaluate(or(ref("meta.ENOENT"), null()), tree), ShouldBeNil)
		})
	})

	Convey("Expression Reduction Algorithm", t, func() {
		var orig, final *Expr
		var err error

		Convey("ignores singleton expression", func() {
			orig = str("string")
			final, err = orig.Reduce()
			So(err, ShouldBeNil)
			exprOk(final, orig)

			orig = null()
			final, err = orig.Reduce()
			So(err, ShouldBeNil)
			exprOk(final, orig)

			orig = ref("meta.key")
			final, err = orig.Reduce()
			So(err, ShouldBeNil)
			exprOk(final, orig)
		})

		Convey("handles normal alternates that terminated in a literal", func() {
			orig = or(ref("a.b.c"), str("default"))
			final, err = orig.Reduce()
			So(err, ShouldBeNil)
			exprOk(final, orig)
		})

		Convey("throws errors (warnings) for unreachable alternates", func() {
			orig = or(null(), str("ignored"))
			final, err = orig.Reduce()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, `literal nil short-circuits expression (nil || "ignored")`)
			exprOk(final, null())

			orig = or(ref("some.key"), or(str("default"), ref("ignored.key")))
			final, err = orig.Reduce()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, `literal "default" short-circuits expression (some.key || "default" || ignored.key)`)
			exprOk(final, or(ref("some.key"), str("default")))

			orig = or(or(ref("some.key"), str("default")), ref("ignored.key"))
			final, err = orig.Reduce()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, `literal "default" short-circuits expression (some.key || "default" || ignored.key)`)
			exprOk(final, or(ref("some.key"), str("default")))
		})
	})

	Convey("File Operator", t, func() {
		op := FileOperator{}
		ev := &Evaluator{
			Tree: YAML(
				`meta:
  sample_file: assets/file_operator/sample.txt
`),
		}
		basedir, _ := os.Getwd()

		Convey("can read a direct file", func() {
			r, err := op.Run(ev, []*Expr{
				str("assets/file_operator/test.txt"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "This is a test\n")
		})

		Convey("can read a file from a reference", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.sample_file"),
			})

			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)

			content, err := ioutil.ReadFile("assets/file_operator/sample.txt")
			So(r.Value.(string), ShouldEqual, string(content))
		})

		Convey("can read a file relative to a specified base path", func() {
			os.Setenv("SPRUCE_FILE_BASE_PATH", filepath.Join(basedir, "assets/file_operator"))
			r, err := op.Run(ev, []*Expr{
				str("test.txt"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "This is a test\n")
		})

		if _, err := os.Stat("/etc/hosts"); err == nil {
			Convey("can read an absolute path", func() {
				os.Setenv("SPRUCE_FILE_BASE_PATH", filepath.Join(basedir, "assets/file_operator"))
				r, err := op.Run(ev, []*Expr{
					str("/etc/hosts"),
				})
				So(err, ShouldBeNil)
				So(r, ShouldNotBeNil)

				So(r.Type, ShouldEqual, Replace)

				content, err := ioutil.ReadFile("/etc/hosts")
				So(r.Value.(string), ShouldEqual, string(content))
			})
		}

		Convey("can handle a missing file", func() {
			r, err := op.Run(ev, []*Expr{
				str("no_one_should_ever_name_a_file_that_doesnt_exist_this_name"),
			})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

	})

	Convey("Grab Operator", t, func() {
		op := GrabOperator{}
		ev := &Evaluator{
			Tree: YAML(
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

		Convey("can grab a single value", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.subkey.value"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "found it")
		})

		Convey("can grab a single list value", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.lonely"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)

			l, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)

			So(len(l), ShouldEqual, 1)
			So(l[0], ShouldEqual, "one")
		})

		Convey("can grab a multiple lists and flatten them", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.list1"),
				ref("key.lonely"),
				ref("key.list2"),
				ref("key.lonely.0"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)

			l, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)

			So(len(l), ShouldEqual, 6)
			So(l[0], ShouldEqual, "first")
			So(l[1], ShouldEqual, "second")
			So(l[2], ShouldEqual, "one")
			So(l[3], ShouldEqual, "third")
			So(l[4], ShouldEqual, "fourth")
			So(l[5], ShouldEqual, "one")
		})

		Convey("can grab multiple values", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.subkey.value"),
				ref("key.subkey.other"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 2)
			So(v[0], ShouldEqual, "found it")
			So(v[1], ShouldEqual, "value 2")
		})

		Convey("flattens constituent arrays", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.list2"),
				ref("key.list1"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 4)
			So(v[0], ShouldEqual, "third")
			So(v[1], ShouldEqual, "fourth")
			So(v[2], ShouldEqual, "first")
			So(v[3], ShouldEqual, "second")
		})

		Convey("throws errors for missing arguments", func() {
			_, err := op.Run(ev, []*Expr{})
			So(err, ShouldNotBeNil)
		})

		Convey("throws errors for dangling references", func() {
			_, err := op.Run(ev, []*Expr{
				ref("key.that.does.not.exist"),
			})
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Environment Variable Resolution (via grab)", t, func() {
		op := GrabOperator{}
		ev := &Evaluator{}
		os.Setenv("GRAB_ONE", "one")
		os.Setenv("GRAB_TWO", "two")
		os.Setenv("GRAB_NOT", "")
		os.Setenv("GRAB_BOOL", "true")
		os.Setenv("GRAB_MULTILINE", `line1

line3
line4`)

		Convey("can grab a single environment value", func() {
			r, err := op.Run(ev, []*Expr{
				env("GRAB_ONE"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "one")
		})

		Convey("tries alternates until it finds a set environment variable", func() {
			r, err := op.Run(ev, []*Expr{
				or(env("GRAB_THREE"), or(env("GRAB_TWO"), env("GRAB_ONE"))),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "two")
		})

		Convey("unmarshalls variable contents", func() {
			r, err := op.Run(ev, []*Expr{
				env("GRAB_BOOL"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(bool), ShouldEqual, true)
		})

		Convey("does not unmarshall string-only variables", func() {
			r, err := op.Run(ev, []*Expr{
				env("GRAB_MULTILINE"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, `line1

line3
line4`)
		})

		Convey("throws errors for unset environment variables", func() {
			_, err := op.Run(ev, []*Expr{
				env("GRAB_NOT"),
			})
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Concat Operator", t, func() {
		op := ConcatOperator{}
		ev := &Evaluator{
			Tree: YAML(
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

		Convey("can concat a single value", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.subkey.value"),
				ref("key.list1.0"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "found itfirst")
		})

		Convey("can concat a literal values", func() {
			r, err := op.Run(ev, []*Expr{
				str("a literal "),
				str("value"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "a literal value")
		})

		Convey("can concat multiple values", func() {
			r, err := op.Run(ev, []*Expr{
				str("I "),
				ref("key.subkey.value"),
				str("!"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "I found it!")
		})

		Convey("can concat integer literals", func() {
			r, err := op.Run(ev, []*Expr{
				str("the answer = "),
				ref("douglas.adams"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "the answer = 42")
		})

		Convey("can concat float literals", func() {
			r, err := op.Run(ev, []*Expr{
				ref("math.PI"),
				str(" is PI"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "3.14159 is PI")
		})

		Convey("throws errors for missing arguments", func() {
			_, err := op.Run(ev, []*Expr{})
			So(err, ShouldNotBeNil)

			_, err = op.Run(ev, []*Expr{str("one")})
			So(err, ShouldNotBeNil)
		})

		Convey("throws errors for dangling references", func() {
			_, err := op.Run(ev, []*Expr{
				ref("key.that.does.not.exist"),
				str("string"),
			})
			So(err, ShouldNotBeNil)
		})
	})

	Convey("static_ips Operator", t, func() {
		op := StaticIPOperator{}
		Reset(func() {
			UsedIPs = map[string]string{}
		})

		Convey("can resolve valid networks inside of job contexts", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 3)
			So(v[0], ShouldEqual, "10.0.0.5")
			So(v[1], ShouldEqual, "10.0.0.6")
			So(v[2], ShouldEqual, "10.0.0.7")
		})
		Convey("works with new style bosh manifests", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 2)
			So(v[0], ShouldEqual, "10.0.0.5")
			So(v[1], ShouldEqual, "10.0.0.6")
		})

		Convey("works with multiple subnets", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 4)
			So(v[0], ShouldEqual, "10.0.0.2")
			So(v[1], ShouldEqual, "10.0.0.3")
			So(v[2], ShouldEqual, "10.0.1.5")
			So(v[3], ShouldEqual, "10.0.1.6")
		})

		Convey("works with multiple subnets with an availability zone", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 4)
			So(v[0], ShouldEqual, "10.0.0.2")
			So(v[1], ShouldEqual, "10.0.0.3")
			So(v[2], ShouldEqual, "10.0.1.5")
			So(v[3], ShouldEqual, "10.0.1.6")
		})

		Convey("works with instance_group availability zones", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 3)
			So(v[0], ShouldEqual, "10.0.1.5")
			So(v[1], ShouldEqual, "10.0.1.6")
			So(v[2], ShouldEqual, "10.0.1.7")
		})

		Convey("works with directly specified availability zones", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 6)
			So(v[0], ShouldEqual, "10.0.2.7")
			So(v[1], ShouldEqual, "10.0.0.2")
			So(v[2], ShouldEqual, "10.0.0.4")
			So(v[3], ShouldEqual, "10.0.2.8")
			So(v[4], ShouldEqual, "10.0.0.3")
			So(v[5], ShouldEqual, "10.0.2.10")
		})

		Convey("throws an error if an unknown availability zone is used in operator", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if offset for an availability zone is out of bounds", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if an instance_group availability zone is not found in subnets", func() {
			ev := &Evaluator{
				Here: cursor("instance_groups.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("can resolve valid large networks inside of job contexts", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
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

				// 1st octet rollover testing disabled due to improve speed.
				// but verified working on 11/30/2015 - gfranks
				//				num(16777215), // 2^24 - 1
				//				num(16777216), // 2^24
				//				num(16777217), // 2^24 + 1
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			v, ok := r.Value.([]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 7)
			So(v[0], ShouldEqual, "10.0.0.0")
			So(v[1], ShouldEqual, "10.0.0.255")
			So(v[2], ShouldEqual, "10.0.1.0") //  3rd octet rollover
			So(v[3], ShouldEqual, "10.0.1.1")

			So(v[4], ShouldEqual, "10.0.255.255")
			So(v[5], ShouldEqual, "10.1.0.0") //  2nd octet rollover
			So(v[6], ShouldEqual, "10.1.0.1")

			// 1st octet rollover testing disabled due to improve speed.
			// but verified working on 11/30/2015 - gfranks
			//			So(v[7], ShouldEqual, "10.255.255.255")
			//			So(v[8], ShouldEqual, "11.0.0.0") //  1st octet rollover
			//			So(v[9], ShouldEqual, "11.0.0.1")
		})

		Convey("throws an error if no job name is specified", func() {
			ev := &Evaluator{
				Here: cursor("jobs.0.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if no job instances specified", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if job instances is not a number", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if job has no network name", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if network has no subnets key", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if network has no subnets", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if network has no static ranges", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if network has malformed static range array(s)", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if network static range has malformed IP addresses", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if the static address pool is too small", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("throws an error if the address pool ends before it starts", func() {
			ev := &Evaluator{
				Here: cursor("jobs.job1.networks.0.static_ips"),
				Tree: YAML(
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
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "ends before it starts")
			So(r, ShouldBeNil)
		})
	})

	Convey("inject Operator", t, func() {
		op := InjectOperator{}
		ev := &Evaluator{
			Tree: YAML(
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

		Convey("can inject a single sub-map", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.subkey"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Inject)
			v, ok := r.Value.(map[interface{}]interface{})
			So(ok, ShouldBeTrue)
			So(v["value"], ShouldEqual, "found it")
			So(v["other"], ShouldEqual, "value 2")
		})

		Convey("can inject multiple sub-maps", func() {
			r, err := op.Run(ev, []*Expr{
				ref("key.subkey"),
				ref("key.subkey2"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Inject)
			v, ok := r.Value.(map[interface{}]interface{})
			So(ok, ShouldBeTrue)
			So(len(v), ShouldEqual, 3)
			So(v["value"], ShouldEqual, "overridden")
			So(v["other"], ShouldEqual, "value 2")
			So(v["third"], ShouldEqual, "trois")
		})

		Convey("handles non-existent references", func() {
			_, err := op.Run(ev, []*Expr{
				ref("key.subkey"),
				ref("key.subkey2"),
				ref("key.subkey2.ENOENT"),
			})
			So(err, ShouldNotBeNil)
		})

		Convey("throws an error when trying to inject a scalar", func() {
			_, err := op.Run(ev, []*Expr{
				ref("key.subkey.value"),
			})
			So(err, ShouldNotBeNil)
		})

		Convey("throws an error when trying to inject a list", func() {
			_, err := op.Run(ev, []*Expr{
				ref("key.list1"),
			})
			So(err, ShouldNotBeNil)
		})
	})

	Convey("param Operator", t, func() {
		op := ParamOperator{}
		ev := &Evaluator{}

		Convey("always causes an error", func() {
			r, err := op.Run(ev, []*Expr{
				str("this is the error"),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "this is the error")
			So(r, ShouldBeNil)
		})
	})

	Convey("Join Operator", t, func() {
		op := JoinOperator{}
		ev := &Evaluator{
			Tree: YAML(
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

		Convey("can join a simple list", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.authorities"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "password.write,clients.write,clients.read,scim.write,scim.read,uaa.admin,clients.secret")
		})

		Convey("can join multiple lists", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.authorities"),
				ref("meta.secondlist"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "password.write,clients.write,clients.read,scim.write,scim.read,uaa.admin,clients.secret,admin.write,admin.read")
		})

		Convey("can join an empty list", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.emptylist"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "")
		})

		Convey("can join string literals", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				str("password.write"),
				str("clients.write"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "password.write,clients.write")
		})

		Convey("can join integer literals", func() {
			r, err := op.Run(ev, []*Expr{
				str(":"),
				num(4), num(8), num(15),
				num(16), num(23), num(42),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "4:8:15:16:23:42")
		})

		Convey("can join referenced string entry", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.somestanza.foo"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "bar")
		})

		Convey("can join referenced string entries", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.somestanza.foo"),
				ref("meta.somestanza.wom"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "bar,bat")
		})

		Convey("can join multiple referenced entries", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.authorities"),
				ref("meta.somestanza.foo"),
				ref("meta.somestanza.wom"),
				str("ending"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "password.write,clients.write,clients.read,scim.write,scim.read,uaa.admin,clients.secret,bar,bat,ending")
		})

		Convey("throws an error when there are no arguments", func() {
			r, err := op.Run(ev, []*Expr{})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no arguments specified")
			So(r, ShouldBeNil)
		})

		Convey("throws an error when there are too few arguments", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "too few arguments supplied")
			So(r, ShouldBeNil)
		})

		Convey("throws an error when separator argument is not a literal", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.emptylist"),
				ref("meta.authorities"),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "join operator only accepts literal argument for the separator")
			So(r, ShouldBeNil)
		})

		Convey("throws an error when referenced entry is not a list or literal", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.somestanza"),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "referenced entry is not a list or string")
			So(r, ShouldBeNil)
		})

		Convey("throws an error when referenced list contains non-string entries", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.anotherkey"),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "is not compatible for")
			So(r, ShouldBeNil)
		})

		Convey("throws an error when there are unresolvable references", func() {
			r, err := op.Run(ev, []*Expr{
				str(","),
				ref("meta.non-existent"),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Unable to resolve")
			So(r, ShouldBeNil)
		})

		Convey("calculates dependencies correctly", func() {

			//TODO: Move this to a higher scope when more dependencies tests are added
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
				//make an array so we can give some coherent output on error
				for _, dep := range deps {
					//Pass through tree so that tests can tolerate changes to the cursor lib
					actualPaths = append(actualPaths, dep.String())
				}
				//sort and compare
				sort.Strings(actualPaths)
				sort.Strings(paths)
				match := reflect.DeepEqual(actualPaths, paths)
				//give result
				if !match {
					return fmt.Sprintf("actual: %+v\n expected: %+v", actualPaths, paths)
				}
				return ""
			}

			Convey("with a single list", func() {
				deps := op.Dependencies(ev, []*Expr{
					str(" "),
					ref("meta.secondlist"),
				},
					nil,
					nil)
				So(deps, shouldHaveDeps, "meta.secondlist.[0]", "meta.secondlist.[1]")
			})

			Convey("with multiple lists", func() {
				deps := op.Dependencies(ev, []*Expr{
					str(" "),
					ref("meta.authorities"),
					ref("meta.secondlist"),
				},
					nil,
					nil)
				So(deps, shouldHaveDeps, "meta.authorities.[0]", "meta.authorities.[1]",
					"meta.authorities.[2]", "meta.authorities.[3]", "meta.authorities.[4]",
					"meta.authorities.[5]", "meta.authorities.[6]",
					"meta.secondlist.[0]", "meta.secondlist.[1]")
			})

			Convey("with a reference string", func() {
				deps := op.Dependencies(ev, []*Expr{
					str(" "),
					ref("meta.somestanza.foo"),
				},
					nil,
					nil)
				So(deps, shouldHaveDeps, "meta.somestanza.foo")
			})

			Convey("with multiple reference strings", func() {
				deps := op.Dependencies(ev, []*Expr{
					str(" "),
					ref("meta.somestanza.foo"),
					ref("meta.somestanza.wom"),
				},
					nil,
					nil)
				So(deps, shouldHaveDeps, "meta.somestanza.foo", "meta.somestanza.wom")
			})

			Convey("with a reference string and a list", func() {
				deps := op.Dependencies(ev, []*Expr{
					str(" "),
					ref("meta.somestanza.foo"),
					ref("meta.secondlist"),
				},
					nil,
					nil)
				So(deps, shouldHaveDeps, "meta.somestanza.foo", "meta.secondlist.[0]",
					"meta.secondlist.[1]")
			})

			Convey("with a literal string", func() {
				deps := op.Dependencies(ev, []*Expr{
					str(" "),
					str("literally literal"),
				},
					nil,
					nil)
				So(deps, shouldHaveDeps)
			})

			Convey("with a literal string and a reference string", func() {
				deps := op.Dependencies(ev, []*Expr{
					str(" "),
					str("beep"),
					ref("meta.somestanza.foo"),
				},
					nil,
					nil)
				So(deps, shouldHaveDeps, "meta.somestanza.foo")
			})
		})
	})

	Convey("empty operator", t, func() {
		op := EmptyOperator{}
		ev := &Evaluator{
			Tree: YAML(
				`---
meta:
  authorities: meep
`),
		}

		//These three are with unquoted arguments (references)
		Convey("can replace with a hash", func() {
			r, err := op.Run(ev, []*Expr{
				ref("hash"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			val, isHash := r.Value.(map[string]interface{})
			So(isHash, ShouldBeTrue)
			So(val, ShouldResemble, map[string]interface{}{})
		})

		Convey("can replace with an array", func() {
			r, err := op.Run(ev, []*Expr{
				ref("array"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			val, isArray := r.Value.([]interface{})
			So(isArray, ShouldBeTrue)
			So(val, ShouldResemble, []interface{}{})
		})

		Convey("can replace with an empty string", func() {
			r, err := op.Run(ev, []*Expr{
				ref("string"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			val, isString := r.Value.(string)
			So(isString, ShouldBeTrue)
			So(val, ShouldEqual, "")
		})

		Convey("throws an error for unrecognized types", func() {
			r, err := op.Run(ev, []*Expr{
				ref("void"),
			})
			So(r, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("works with string literals", func() {
			r, err := op.Run(ev, []*Expr{
				str("hash"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)
			So(r.Type, ShouldEqual, Replace)
			val, isHash := r.Value.(map[string]interface{})
			So(isHash, ShouldBeTrue)
			So(val, ShouldResemble, map[string]interface{}{})
		})

		Convey("throws an error with no args", func() {
			r, err := op.Run(ev, []*Expr{})
			So(r, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

		Convey("throws an error with too many args", func() {
			r, err := op.Run(ev, []*Expr{
				ref("hash"),
				ref("array"),
			})
			So(r, ShouldBeNil)
			So(err, ShouldNotBeNil)
		})

	})

	Convey("ips Operator", t, func() {
		op := IpsOperator{}
		ev := &Evaluator{
			Tree: YAML(
				`meta:
  base_network: 1.2.3.4/24
  base_ip: 1.2.3.4
  index: 20
  negative_index: -20
  count: 2
`),
		}

		Convey("can build a single IP based on refs (CIDR)", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.base_network"),
				ref("meta.index"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "1.2.3.20")
		})

		Convey("can build a single IP based on refs (IP)", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.base_ip"),
				ref("meta.index"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "1.2.3.24")
		})

		Convey("can build a single IP based on literals", func() {
			r, err := op.Run(ev, []*Expr{
				str("1.2.3.4/24"),
				num(20),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "1.2.3.20")
		})

		Convey("can build a list of IP's based on references (CIDR)", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.base_network"),
				ref("meta.index"),
				ref("meta.count"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.([]interface{})[0].(string), ShouldEqual, "1.2.3.20")
			So(r.Value.([]interface{})[1].(string), ShouldEqual, "1.2.3.21")
			So(len(r.Value.([]interface{})), ShouldEqual, 2)
		})

		Convey("can build a list of IP's based on literals", func() {
			r, err := op.Run(ev, []*Expr{
				str("1.2.3.4/24"),
				num(20),
				num(2),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.([]interface{})[0].(string), ShouldEqual, "1.2.3.20")
			So(r.Value.([]interface{})[1].(string), ShouldEqual, "1.2.3.21")
			So(len(r.Value.([]interface{})), ShouldEqual, 2)
		})

		Convey("can build a list of IP's based on references (IP)", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.base_ip"),
				ref("meta.index"),
				ref("meta.count"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.([]interface{})[0].(string), ShouldEqual, "1.2.3.24")
			So(r.Value.([]interface{})[1].(string), ShouldEqual, "1.2.3.25")
			So(len(r.Value.([]interface{})), ShouldEqual, 2)
		})

		Convey("can build a list of IP's using negative index (IP)", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.base_ip"),
				ref("meta.negative_index"),
				ref("meta.count"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.([]interface{})[0].(string), ShouldEqual, "1.2.2.240")
			So(r.Value.([]interface{})[1].(string), ShouldEqual, "1.2.2.241")
			So(len(r.Value.([]interface{})), ShouldEqual, 2)
		})

		Convey("can build a list of IP's using negative index (CIDR)", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.base_network"),
				ref("meta.negative_index"),
				ref("meta.count"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.([]interface{})[0].(string), ShouldEqual, "1.2.3.236")
			So(r.Value.([]interface{})[1].(string), ShouldEqual, "1.2.3.237")
			So(len(r.Value.([]interface{})), ShouldEqual, 2)
		})

		Convey("bails out if index is outside CIDR size", func() {
			r, err := op.Run(ev, []*Expr{
				str("192.168.1.16/29"),
				num(100),
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Start index 100 exceeds size of subnet 192.168.1.16/29")
			So(r, ShouldBeNil)
		})

		Convey("bails out if count would go outside CIDR size", func() {
			r, err := op.Run(ev, []*Expr{
				str("192.168.1.16/29"),
				num(-1),
				num(3),
			})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Start index 7 and count 3 would exceed size of subnet 192.168.1.16/29")
			So(r, ShouldBeNil)
		})
	})

	Convey("Base64 Operator", t, func() {
		op := Base64Operator{}
		ev := &Evaluator{
			Tree: YAML(
				`meta:
  sample: "Sample Text To Base64 Encode From Reference"
`),
		}

		Convey("can encode a string literal", func() {
			r, err := op.Run(ev, []*Expr{
				str("Sample Text To Base64 Encode"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "U2FtcGxlIFRleHQgVG8gQmFzZTY0IEVuY29kZQ==")
		})

		Convey("can encode from a reference", func() {
			r, err := op.Run(ev, []*Expr{
				ref("meta.sample"),
			})

			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)

			So(r.Value.(string), ShouldEqual, "U2FtcGxlIFRleHQgVG8gQmFzZTY0IEVuY29kZSBGcm9tIFJlZmVyZW5jZQ==")
		})

		Convey("can handle non string scalar input", func() {
			r, err := op.Run(ev, []*Expr{
				str("one"), num(1),
			})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

		Convey("can handle non string scalar input (i.e numbers)", func() {
			r, err := op.Run(ev, []*Expr{
				num(1),
			})
			So(err, ShouldNotBeNil)
			So(r, ShouldBeNil)
		})

	})

	Convey("awsparam/awssecret operator", t, func() {
		op := AwsOperator{variant: "awsparam"}
		ev := &Evaluator{
			Tree: YAML(`{ "testval": "test", "testmap": {}, "testarr": [] }`),
			Here: &tree.Cursor{},
		}
		mock := &mockedAws{}

		var ssmKey string
		var ssmRet string
		var ssmErr error

		mock.MockGetParameter = func(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
			ssmKey = aws.StringValue(in.Name)
			return &ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					Value: aws.String(ssmRet),
				},
			}, ssmErr
		}

		parameterstoreClient = mock
		secretsManagerClient = mock

		Convey("in shared logic", func() {
			Convey("should return error if no key given", func() {
				_, err := op.Run(ev, []*Expr{})
				So(err.Error(), ShouldContainSubstring, "awsparam operator requires at least one argument")
			})

			Convey("should concatenate args", func() {
				_, err := op.Run(ev, []*Expr{num(1), num(2), num(3)})
				So(err, ShouldBeNil)
				So(ssmKey, ShouldEqual, "123")
			})

			Convey("should resolve references", func() {
				_, err := op.Run(ev, []*Expr{num(1), num(2), ref("testval")})
				So(err, ShouldBeNil)
				So(ssmKey, ShouldEqual, "12test")
			})

			Convey("should not allow references to maps", func() {
				_, err := op.Run(ev, []*Expr{num(1), num(2), ref("testmap")})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "$.testmap is a map; only scalars are supported here")
			})

			Convey("should not allow references to arrays", func() {
				_, err := op.Run(ev, []*Expr{num(1), num(2), ref("testarr")})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "$.testarr is a list; only scalars are supported here")
			})

			Convey("without key", func() {
				ssmRet = "testx"
				r, err := op.Run(ev, []*Expr{str("val1")})
				So(err, ShouldBeNil)
				So(r.Type, ShouldEqual, Replace)
				So(r.Value.(string), ShouldEqual, "testx")
			})

			Convey("with key", func() {
				Convey("should parse subkey and extract if provided", func() {
					ssmRet = `{ "key": "val" }`
					r, err := op.Run(ev, []*Expr{str("val2?key=key")})
					So(err, ShouldBeNil)
					So(r.Type, ShouldEqual, Replace)
					So(r.Value.(string), ShouldEqual, "val")
				})

				Convey("should error if document not valid yaml / json", func() {
					ssmRet = `key: {`
					_, err := op.Run(ev, []*Expr{str("val3?key=key")})
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "$.val3 error extracting key: yaml: line 1: did not find expected node content")
				})

				Convey("should error if subkey invalid", func() {
					ssmRet = `key: {}`
					_, err := op.Run(ev, []*Expr{str("val4?key=noexist")})
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "$.val4 invalid key 'noexist'")
				})
			})

			Convey("should not call AWS API if SkipAws true", func() {
				SkipAws = true
				count := 0
				mock.MockGetParameter = func(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
					count = count + 1
					return &ssm.GetParameterOutput{
						Parameter: &ssm.Parameter{
							Value: aws.String(""),
						},
					}, nil
				}
				_, err := op.Run(ev, []*Expr{str("skipaws")})
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 0)
				SkipAws = false
			})
		})

		Convey("awsparam", func() {
			Convey("should cache lookups", func() {
				count := 0
				mock.MockGetParameter = func(in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
					count = count + 1
					return &ssm.GetParameterOutput{
						Parameter: &ssm.Parameter{
							Value: aws.String(""),
						},
					}, nil
				}
				_, err := op.Run(ev, []*Expr{str("val5")})
				So(err, ShouldBeNil)
				_, err = op.Run(ev, []*Expr{str("val5")})
				So(err, ShouldBeNil)

				So(count, ShouldEqual, 1)
			})
		})

		Convey("awssecret", func() {
			op = AwsOperator{variant: "awssecret"}
			Convey("should cache lookups", func() {
				count := 0
				mock.MockGetSecretValue = func(in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
					count = count + 1
					return &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String(""),
					}, nil
				}
				_, err := op.Run(ev, []*Expr{str("val5")})
				So(err, ShouldBeNil)
				_, err = op.Run(ev, []*Expr{str("val5")})
				So(err, ShouldBeNil)

				So(count, ShouldEqual, 1)
			})

			Convey("should use stage if provided", func() {
				stage := ""
				mock.MockGetSecretValue = func(in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
					stage = aws.StringValue(in.VersionStage)
					return &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String(""),
					}, nil
				}
				_, err := op.Run(ev, []*Expr{str("val6?stage=test")})
				So(err, ShouldBeNil)

				So(stage, ShouldEqual, "test")
			})

			Convey("should use version if provided", func() {
				version := ""
				mock.MockGetSecretValue = func(in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
					version = aws.StringValue(in.VersionId)
					return &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String(""),
					}, nil
				}
				_, err := op.Run(ev, []*Expr{str("val7?version=test")})
				So(err, ShouldBeNil)

				So(version, ShouldEqual, "test")
			})
		})
	})
}
