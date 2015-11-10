package main

import (
	"github.com/geofffranks/simpleyaml" // FIXME: switch back to smallfish/simpleyaml after https://github.com/smallfish/simpleyaml/pull/1 is merged
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestEvaluator(t *testing.T) {
	YAML := func(s string) map[interface{}]interface{} {
		y, err := simpleyaml.NewYaml([]byte(s))
		So(err, ShouldBeNil)

		data, err := y.Map()
		So(err, ShouldBeNil)

		return data
	}

	Convey("Evaluator", t, func() {
		Convey("Data Flow", func() {
			Convey("Generates a sequential list of operator calls, in order", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  foo: FOO
  bar:  (( grab meta.foo ))
  baz:  (( grab meta.bar ))
  quux: (( grab meta.baz ))
  boz:  (( grab meta.quux ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)
				So(ev.DataOps, ShouldNotBeNil)

				// expect: meta.bar   (( grab meta.foo ))
				//         meta.baz   (( grab meta.bar ))
				//         meta.quux  (( grab meta.baz ))
				//         meta.boz   (( grab meta.quux ))
				So(len(ev.DataOps), ShouldEqual, 4)

				So(ev.DataOps[0].where.String(), ShouldEqual, "meta.bar")
				So(ev.DataOps[0].src, ShouldEqual, "(( grab meta.foo ))")

				So(ev.DataOps[1].where.String(), ShouldEqual, "meta.baz")
				So(ev.DataOps[1].src, ShouldEqual, "(( grab meta.bar ))")

				So(ev.DataOps[2].where.String(), ShouldEqual, "meta.quux")
				So(ev.DataOps[2].src, ShouldEqual, "(( grab meta.baz ))")

				So(ev.DataOps[3].where.String(), ShouldEqual, "meta.boz")
				So(ev.DataOps[3].src, ShouldEqual, "(( grab meta.quux ))")
			})

			Convey("detects direct (a -> b -> a) cycles in data flow graph", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  bar: (( grab meta.foo ))
  foo: (( grab meta.bar ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldNotBeNil)
			})

			Convey("detects indirect (a -> b -> c -> a) cycles in data flow graph", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  foo: (( grab meta.bar ))
  bar: (( grab meta.baz ))
  baz: (( grab meta.foo ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldNotBeNil)
			})

			Convey("handles data flow regardless of operator type", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  foo: FOO
  bar: (( grab meta.foo ))
  baz: (( grab meta.bar ))
  quux: (( concat "literal:" meta.baz ))
  boz: (( grab meta.quux ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				// expect: meta.bar   (( grab meta.foo ))
				//         meta.baz   (( grab meta.bar ))
				//         meta.quux  (( concat "literal:" meta.baz ))
				//         meta.boz   (( grab meta.quux ))
				So(len(ev.DataOps), ShouldEqual, 4)

				So(ev.DataOps[0].where.String(), ShouldEqual, "meta.bar")
				So(ev.DataOps[0].src, ShouldEqual, "(( grab meta.foo ))")

				So(ev.DataOps[1].where.String(), ShouldEqual, "meta.baz")
				So(ev.DataOps[1].src, ShouldEqual, "(( grab meta.bar ))")

				So(ev.DataOps[2].where.String(), ShouldEqual, "meta.quux")
				So(ev.DataOps[2].src, ShouldEqual, `(( concat "literal:" meta.baz ))`)

				So(ev.DataOps[3].where.String(), ShouldEqual, "meta.boz")
				So(ev.DataOps[3].src, ShouldEqual, "(( grab meta.quux ))")
			})

			Convey("handles data flow for operators in lists", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  - FOO
  - (( grab meta.0 ))
  - (( grab meta.1 ))
  - (( concat "literal:" meta.2 ))
  - (( grab meta.3 ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				// expect: meta.1  (( grab meta.0 ))
				//         meta.2  (( grab meta.1 ))
				//         meta.3  (( concat "literal:" meta.2 ))
				//         meta.4  (( grab meta.3 ))
				So(len(ev.DataOps), ShouldEqual, 4)

				So(ev.DataOps[0].where.String(), ShouldEqual, "meta.1")
				So(ev.DataOps[0].src, ShouldEqual, "(( grab meta.0 ))")

				So(ev.DataOps[1].where.String(), ShouldEqual, "meta.2")
				So(ev.DataOps[1].src, ShouldEqual, "(( grab meta.1 ))")

				So(ev.DataOps[2].where.String(), ShouldEqual, "meta.3")
				So(ev.DataOps[2].src, ShouldEqual, `(( concat "literal:" meta.2 ))`)

				So(ev.DataOps[3].where.String(), ShouldEqual, "meta.4")
				So(ev.DataOps[3].src, ShouldEqual, "(( grab meta.3 ))")
			})

			Convey("handles deep copy in data flow graph", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  first: [ a, b, c ]
  second: (( grab meta.first ))
  third:  (( grab meta.second ))
  gotcha: (( grab meta.third.0 ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				// expect: meta.second (( grab meta.first ))
				//         meta.third  (( grab meta.second ))
				//         meta.gotcha (( grab meta.third.0 ))
				//
				//   (the key point here is that meta.third.0 doesn't
				//    exist in the tree until we start evaluating, but
				//    we still need to get the order correct; we should
				//    have a dep on meta.third, and hope that run-time
				//    resolution puts an array there for us to find...)
				//
				So(len(ev.DataOps), ShouldEqual, 3)

				So(ev.DataOps[0].where.String(), ShouldEqual, "meta.second")
				So(ev.DataOps[0].src, ShouldEqual, "(( grab meta.first ))")

				So(ev.DataOps[1].where.String(), ShouldEqual, "meta.third")
				So(ev.DataOps[1].src, ShouldEqual, "(( grab meta.second ))")

				So(ev.DataOps[2].where.String(), ShouldEqual, "meta.gotcha")
				So(ev.DataOps[2].src, ShouldEqual, "(( grab meta.third.0 ))")
			})

			Convey("handles implicit static_ip dependency on jobs.*.networks.*.name", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  environment: prod
  size: 4
networks:
  - name: sandbox
    subnets:
    - static: [ 10.2.0.5 - 10.2.0.10 ]
  - name: prod
    subnets:
    - static: [ 10.0.0.5 - 10.0.0.100 ]
jobs:
  - name: job1
    instances: 4
    networks:
      - name: (( grab meta.environment ))
        static_ips: (( static_ips 1 2 3 4 ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				// expect: jobs.0.networks.0.name        (( grab meta.environment ))
				//         jobs.0.networks.0.static_ips  (( static_ips 1 2 3 4 ))
				So(len(ev.DataOps), ShouldEqual, 2)

				So(ev.DataOps[0].where.String(), ShouldEqual, "jobs.0.networks.0.name")
				So(ev.DataOps[0].src, ShouldEqual, "(( grab meta.environment ))")

				So(ev.DataOps[1].where.String(), ShouldEqual, "jobs.0.networks.0.static_ips")
				So(ev.DataOps[1].src, ShouldEqual, "(( static_ips 1 2 3 4 ))")
			})

			Convey("handles implicit static_ip dependency on networks.*.name", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  net: real
  environment: prod
  size: 4
networks:
  - name: (( concat meta.net "-prod" ))
    subnets:
    - static: [ 10.0.0.5 - 10.0.0.100 ]
jobs:
  - name: job1
    instances: 4
    networks:
      - name: prod-net # must be literal to avoid non-determinism
        static_ips: (( static_ips 1 2 3 4 ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				// expect: networks.0.name               (( concat meta.net "-prod" ))
				//         jobs.0.networks.0.static_ips  (( static_ips 1 2 3 4 ))
				So(len(ev.DataOps), ShouldEqual, 2)

				So(ev.DataOps[0].where.String(), ShouldEqual, "networks.0.name")
				So(ev.DataOps[0].src, ShouldEqual, `(( concat meta.net "-prod" ))`)

				So(ev.DataOps[1].where.String(), ShouldEqual, "jobs.0.networks.0.static_ips")
				So(ev.DataOps[1].src, ShouldEqual, "(( static_ips 1 2 3 4 ))")
			})

			Convey("handles dependency on static_ips() calls via (( grab )) calls", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
networks:
  - name: net1
    subnets:
    - static: [ 10.0.0.5 - 10.0.0.100 ]

jobs:
  - name: job1
    instances: 4
    networks:
      - name: net1
        static_ips: (( static_ips 1 2 3 4 ))

properties:
  job_ips: (( grab jobs.job1.networks.net1.static_ips ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)
				So(ev.DataOps, ShouldNotBeNil)

				// expect: jobs.0.networks.0.static_ips   (( static_ips 1 2 3 4 ))
				//         properties.job_ips             (( grab jobs.job1.networks.net1.static_ips ))
				So(len(ev.DataOps), ShouldEqual, 2)

				So(ev.DataOps[0].where.String(), ShouldEqual, "jobs.0.networks.0.static_ips")
				So(ev.DataOps[0].src, ShouldEqual, "(( static_ips 1 2 3 4 ))")

				So(ev.DataOps[1].where.String(), ShouldEqual, "properties.job_ips")
				So(ev.DataOps[1].src, ShouldEqual, "(( grab jobs.job1.networks.net1.static_ips ))")
			})

			Convey("handles implicit deps on sub-tree operations in (( inject ... )) targets", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  template:
    foo: bar
    baz: (( grab meta.template.foo ))

thing:
  <<<: (( inject meta.template ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				So(len(ev.DataOps), ShouldEqual, 2)

				So(ev.DataOps[0].where.String(), ShouldEqual, "meta.template.baz")
				So(ev.DataOps[0].src, ShouldEqual, "(( grab meta.template.foo ))")

				So(ev.DataOps[1].where.String(), ShouldEqual, "thing.<<<")
				So(ev.DataOps[1].src, ShouldEqual, "(( inject meta.template ))")
			})

			Convey("handles inject of an inject of a grab (so meta)", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  template1:
    foo: bar
    baz: (( grab meta.template1.foo ))
  template2:
    <<<: (( inject meta.template1 ))
    xyzzy: nothing happens

thing:
  <<<: (( inject meta.template2 ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				So(len(ev.DataOps), ShouldEqual, 3)

				So(ev.DataOps[0].where.String(), ShouldEqual, "meta.template1.baz")
				So(ev.DataOps[0].src, ShouldEqual, "(( grab meta.template1.foo ))")

				So(ev.DataOps[1].where.String(), ShouldEqual, "meta.template2.<<<")
				So(ev.DataOps[1].src, ShouldEqual, "(( inject meta.template1 ))")

				So(ev.DataOps[2].where.String(), ShouldEqual, "thing.<<<")
				So(ev.DataOps[2].src, ShouldEqual, "(( inject meta.template2 ))")
			})
		})

		Convey("Patching", func() {
			valueIs := func(ev *Evaluator, path string, expect string) {
				c, err := ParseCursor(path)
				So(err, ShouldBeNil)

				v, err := c.ResolveString(ev.Tree)
				So(err, ShouldBeNil)

				So(v, ShouldEqual, expect)
			}
			notPresent := func(ev *Evaluator, path string) {
				c, err := ParseCursor(path)
				So(err, ShouldBeNil)

				_, err = c.ResolveString(ev.Tree)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "could not be found")
			}

			Convey("can handle simple map-based Replace actions", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  domain: sandbox.example.com
  web: (( grab meta.domain ))
urls:
  home: (( concat "http://www." meta.web ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				err = ev.Patch()
				So(err, ShouldBeNil)

				valueIs(ev, "meta.web", "sandbox.example.com")
				valueIs(ev, "urls.home", "http://www.sandbox.example.com")
			})

			Convey("can handle Replacement actions where the new value is a list", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  things:
    - one
    - two
grocery:
  list: (( grab meta.things ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				err = ev.Patch()
				So(err, ShouldBeNil)

				valueIs(ev, "grocery.list.0", "one")
				valueIs(ev, "grocery.list.1", "two")
			})

			Convey("can handle Replacement actions where the call site is a list", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  first:  2nd
  second: 1st
sorted:
  list:
    - (( grab meta.second ))
    - (( grab meta.first ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				err = ev.Patch()
				So(err, ShouldBeNil)

				valueIs(ev, "sorted.list.0", "1st")
				valueIs(ev, "sorted.list.1", "2nd")
			})

			Convey("can handle Replacement actions where the call site is inside of a list", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  prod: production
  sandbox: sb322
boxen:
  - name: www
    env: (( grab meta.prod ))
  - name: wwwtest
    env: (( grab meta.sandbox ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				err = ev.Patch()
				So(err, ShouldBeNil)

				valueIs(ev, "boxen.www.env", "production")
				valueIs(ev, "boxen.wwwtest.env", "sb322")
			})

			Convey("can handle simple Inject actions", func() {
				ev := &Evaluator{
					Tree: YAML(
						`templates:
  www:
    HA: enabled
    DR: disabled
host:
  web1:
    type: www
    <<<: (( inject templates.www ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				err = ev.Patch()
				So(err, ShouldBeNil)

				valueIs(ev, "host.web1.HA", "enabled")
				valueIs(ev, "host.web1.DR", "disabled")
				valueIs(ev, "host.web1.type", "www")
				notPresent(ev, "host.web1.<<<")
			})

			Convey("can handle Inject actions where call site is in a list", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  jobs:
    api:
      template: api
    worker:
      template: worker
    db:
      template: database

jobs:
  - name: api_z1
    <<<: (( inject meta.jobs.api ))
  - name: api_z2
    <<<: (( inject meta.jobs.api ))

  - name: worker_z3
    <<<: (( inject meta.jobs.worker ))

  - name: db_z3
    <<<: (( inject meta.jobs.db ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				err = ev.Patch()
				So(err, ShouldBeNil)

				valueIs(ev, "jobs.api_z1.template", "api")
				valueIs(ev, "jobs.api_z2.template", "api")

				valueIs(ev, "jobs.db_z3.template", "database")

				valueIs(ev, "jobs.worker_z3.template", "worker")
			})

			Convey("preserves call-site keys on conflict in an Inject scenario", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  template:
    foo: FOO
    bar: BAR

example:
  <<<: (( inject meta.template ))
  foo: foooo
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				err = ev.Patch()
				So(err, ShouldBeNil)

				valueIs(ev, "example.foo", "foooo")
				valueIs(ev, "example.bar", "BAR")
			})

			Convey("merges sub-trees common to inject site and injected values", func() {
				ev := &Evaluator{
					Tree: YAML(
						`meta:
  template:
    properties:
      foo: bar
      baz: NOT-OVERRIDDEN

thing:
  <<<: (( inject meta.template ))
  properties:
    bar: baz
    baz: overridden
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				err = ev.Patch()
				So(err, ShouldBeNil)

				valueIs(ev, "thing.properties.bar", "baz")
				valueIs(ev, "thing.properties.baz", "overridden")
				valueIs(ev, "thing.properties.foo", "bar")
			})

			Convey("handles static_ips() call and a subsequent grab", func() {
				ev := &Evaluator{
					Tree: YAML(
						`jobs:
- name: api_z1
  instances: 1
  networks:
  - name: net1
    static_ips: (( static_ips(0, 1, 2) ))

networks:
- name: net1
  subnets:
    - static: [192.168.1.2 - 192.168.1.30]

properties:
  api_servers: (( grab jobs.api_z1.networks.net1.static_ips ))
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				err = ev.Patch()
				So(err, ShouldBeNil)

				valueIs(ev, "jobs.api_z1.networks.net1.static_ips.0", "192.168.1.2")
				valueIs(ev, "properties.api_servers.0", "192.168.1.2")
			})

			Convey("handles allocation conflicts of static IP addresses", func() {
				ev := &Evaluator{
					Tree: YAML(
						`jobs:
- name: api_z1
  instances: 1
  networks:
  - name: net1
    static_ips: (( static_ips(0, 1, 2) ))
- name: api_z2
  instances: 1
  networks:
  - name: net1
    static_ips: (( static_ips(0, 1, 2) ))

networks:
- name: net1
  subnets:
    - static: [192.168.1.2 - 192.168.1.30]
`),
				}

				err := ev.DataFlow()
				So(err, ShouldBeNil)

				err = ev.Patch()
				So(err, ShouldNotBeNil)
			})
		})
	})
}
