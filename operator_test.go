package main

import (
	"github.com/geofffranks/simpleyaml" // FIXME: switch back to smallfish/simpleyaml after https://github.com/smallfish/simpleyaml/pull/1 is merged
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestOperators(t *testing.T) {
	cursor := func(s string) *Cursor {
		c, err := ParseCursor(s)
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

	R := func(s string) *Expr {
		return &Expr{Type: Reference, Reference: cursor(s)}
	}
	Ls := func(s string) *Expr {
		return &Expr{Type: Literal, Literal: s}
	}
	Li := func(v int64) *Expr {
		return &Expr{Type: Literal, Literal: v}
	}
	/*
		O := func(l *Expr, r *Expr) *Expr {
			return &Expr{Type: LogicalOr, Left: l, Right: r}
		}
	*/

	Convey("Parser", t, func() {
		Convey("parses op calls in their entirety", func() {

			opOk := func(code string, name string, args ...*Expr) {
				op, err := ParseOpcall(code)
				So(err, ShouldBeNil)
				So(op, ShouldNotBeNil)

				_, ok := op.op.(NullOperator)
				So(ok, ShouldBeTrue)
				So(op.op.(NullOperator).Missing, ShouldEqual, name)

				var exprOk func(*Expr, *Expr)
				exprOk = func(got *Expr, want *Expr) {
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

				So(len(op.args), ShouldEqual, len(args))
				for i, expect := range args {
					exprOk(op.args[i], expect)
				}
			}

			Convey("handles opcodes with and without arguments", func() {
				opOk(`(( null ))`, "null")
				opOk(`(( null 42 ))`, "null", Li(42))
				opOk(`(( null 1 2 3 4 ))`, "null", Li(1), Li(2), Li(3), Li(4))
			})

			Convey("ignores optional whitespace", func() {
				opOk(`((null))`, "null")
				opOk(`((	null	))`, "null")
				opOk(`((  	null  	))`, "null")

				args := []*Expr{Li(1), Li(2), Li(3)}
				opOk(`((null 1 2 3))`, "null", args...)
				opOk(`((null 1	2	3))`, "null", args...)
				opOk(`((null 1	2	3	))`, "null", args...)
				opOk(`((null 1 	 2 	 3 	 ))`, "null", args...)
			})

			Convey("allows use of commas to separate arguments", func() {
				args := []*Expr{Li(1), Li(2), Li(3)}
				opOk(`((null 1, 2, 3))`, "null", args...)
				opOk(`((null 1,	2,	3))`, "null", args...)
				opOk(`((null 1,	2,	3,	))`, "null", args...)
				opOk(`((null 1 ,	 2 ,	 3 ,	 ))`, "null", args...)
			})

			Convey("allows use of parentheses around arguments", func() {
				args := []*Expr{Li(1), Li(2), Li(3)}
				opOk(`((null(1,2,3)))`, "null", args...)
				opOk(`((null(1, 2, 3) ))`, "null", args...)
				opOk(`((null( 1,	2,	3)))`, "null", args...)
				opOk(`((null (1,	2,	3)	))`, "null", args...)
				opOk(`((null (1 ,	 2 ,	 3)	 ))`, "null", args...)
			})

			Convey("handles string literal arguments", func() {
				opOk(`(( null "string" ))`, "null", Ls("string"))
				opOk(`(( null "string with whitespace" ))`, "null", Ls("string with whitespace"))
				opOk(`(( null "a \"quoted\" string" ))`, "null", Ls(`a "quoted" string`))
				opOk(`(( null "\\escaped" ))`, "null", Ls(`\escaped`))
			})

			Convey("handles reference (cursor) arguments", func() {
				opOk(`(( null x.y.z ))`, "null", R("x.y.z"))
				opOk(`(( null x.[0].z ))`, "null", R("x.0.z"))
				opOk(`(( null x[0].z ))`, "null", R("x.0.z"))
				opOk(`(( null x[0]z ))`, "null", R("x.0.z"))
			})

			Convey("handles mixed collections of argument types", func() {
				opOk(`(( xyzzy "string" x.y.z 42  ))`, "xyzzy", Ls("string"), R("x.y.z"), Li(42))
				opOk(`(( xyzzy("string" x.y.z 42) ))`, "xyzzy", Ls("string"), R("x.y.z"), Li(42))
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

			e = &Expr{
				Type:  LogicalOr,
				Left:  &Expr{Type: Literal, Literal: "first"},
				Right: &Expr{Type: Literal, Literal: "second"},
			}
			So(evaluate(e, tree), ShouldEqual, "first")

			e = &Expr{
				Type:  LogicalOr,
				Left:  &Expr{Type: Reference, Reference: cursor("meta.foo")},
				Right: &Expr{Type: Literal, Literal: "second"},
			}
			So(evaluate(e, tree), ShouldEqual, "FOO")

			e = &Expr{
				Type:  LogicalOr,
				Left:  &Expr{Type: Reference, Reference: cursor("meta.ENOENT")},
				Right: &Expr{Type: Reference, Reference: cursor("meta.foo")},
			}
		})

		Convey("|| operator treats nil as a found value", func() {
			tree = YAML(`---
meta:
  foo: FOO
  bar: BAR
`)

			e = &Expr{
				Type:  LogicalOr,
				Left:  &Expr{Type: Literal, Literal: nil},
				Right: &Expr{Type: Literal, Literal: "second"},
			}
			So(evaluate(e, tree), ShouldBeNil)

			e = &Expr{
				Type:  LogicalOr,
				Left:  &Expr{Type: Reference, Reference: cursor("meta.ENOENT2")},
				Right: &Expr{Type: Literal, Literal: nil},
			}
			So(evaluate(e, tree), ShouldBeNil)
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
				R("key.subkey.value"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "found it")
		})

		Convey("can grab a single list value", func() {
			r, err := op.Run(ev, []*Expr{
				R("key.lonely"),
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
				R("key.list1"),
				R("key.lonely"),
				R("key.list2"),
				R("key.lonely.0"),
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
				R("key.subkey.value"),
				R("key.subkey.other"),
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
				R("key.list2"),
				R("key.list1"),
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
				R("key.that.does.not.exist"),
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
				R("key.subkey.value"),
				R("key.list1.0"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "found itfirst")
		})

		Convey("can concat a literal values", func() {
			r, err := op.Run(ev, []*Expr{
				Ls("a literal "),
				Ls("value"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "a literal value")
		})

		Convey("can concat multiple values", func() {
			r, err := op.Run(ev, []*Expr{
				Ls("I "),
				R("key.subkey.value"),
				Ls("!"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "I found it!")
		})

		Convey("can concat integer literals", func() {
			r, err := op.Run(ev, []*Expr{
				Ls("the answer = "),
				R("douglas.adams"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "the answer = 42")
		})

		Convey("can concat float literals", func() {
			r, err := op.Run(ev, []*Expr{
				R("math.PI"),
				Ls(" is PI"),
			})
			So(err, ShouldBeNil)
			So(r, ShouldNotBeNil)

			So(r.Type, ShouldEqual, Replace)
			So(r.Value.(string), ShouldEqual, "3.14159 is PI")
		})

		Convey("throws errors for missing arguments", func() {
			_, err := op.Run(ev, []*Expr{})
			So(err, ShouldNotBeNil)

			_, err = op.Run(ev, []*Expr{Ls("one")})
			So(err, ShouldNotBeNil)
		})

		Convey("throws errors for dangling references", func() {
			_, err := op.Run(ev, []*Expr{
				R("key.that.does.not.exist"),
				Ls("string"),
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

			r, err := op.Run(ev, []*Expr{Li(0), Li(1), Li(2)})
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

			r, err := op.Run(ev, []*Expr{Li(0), Li(1), Li(2)})
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

			r, err := op.Run(ev, []*Expr{Li(0), Li(1), Li(2)})
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

			r, err := op.Run(ev, []*Expr{Li(0), Li(1), Li(2)})
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

			r, err := op.Run(ev, []*Expr{Li(0), Li(1), Li(2)})
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

			r, err := op.Run(ev, []*Expr{Li(0), Li(1), Li(2)})
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

			r, err := op.Run(ev, []*Expr{Li(0), Li(1), Li(2)})
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

			r, err := op.Run(ev, []*Expr{Li(0), Li(1), Li(2)})
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

			r, err := op.Run(ev, []*Expr{Li(0), Li(1), Li(2)})
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

			r, err := op.Run(ev, []*Expr{Li(0), Li(1), Li(2)})
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

			r, err := op.Run(ev, []*Expr{Li(0), Li(1), Li(2)})
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
				R("key.subkey"),
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
				R("key.subkey"),
				R("key.subkey2"),
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
				R("key.subkey"),
				R("key.subkey2"),
				R("key.subkey2.ENOENT"),
			})
			So(err, ShouldNotBeNil)
		})

		Convey("throws an error when trying to inject a scalar", func() {
			_, err := op.Run(ev, []*Expr{
				R("key.subkey.value"),
			})
			So(err, ShouldNotBeNil)
		})

		Convey("throws an error when trying to inject a list", func() {
			_, err := op.Run(ev, []*Expr{
				R("key.list1"),
			})
			So(err, ShouldNotBeNil)
		})
	})

	Convey("param Operator", t, func() {
		op := ParamOperator{}
		ev := &Evaluator{}

		Convey("always causes an error", func() {
			r, err := op.Run(ev, []*Expr{
				Ls("this is the error"),
			})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "this is the error")
			So(r, ShouldBeNil)
		})
	})
}
