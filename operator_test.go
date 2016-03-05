package main

import (
	"github.com/jhunt/tree"
	"github.com/smallfish/simpleyaml"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

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

			Convey("handles opcodes with and without arguments", func() {
				opOk(`(( null ))`, "null")
				opOk(`(( null 42 ))`, "null", num(42))
				opOk(`(( null 1 2 3 4 ))`, "null", num(1), num(2), num(3), num(4))
			})

			Convey("ignores optional whitespace", func() {
				opOk(`((null))`, "null")
				opOk(`((	null	))`, "null")
				opOk(`((  	null  	))`, "null")

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
}
