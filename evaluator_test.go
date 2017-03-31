package spruce

import (
	"bufio"
	"regexp"
	"strings"
	"testing"

	// Use geofffranks forks to persist the fix in https://github.com/go-yaml/yaml/pull/133/commits
	// Also https://github.com/go-yaml/yaml/pull/195
	"github.com/geofffranks/simpleyaml"
	"github.com/geofffranks/yaml"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEvaluator(t *testing.T) {
	SilenceWarnings(true)
	YAML := func(s string) map[interface{}]interface{} {
		y, err := simpleyaml.NewYaml([]byte(s))
		So(err, ShouldBeNil)

		data, err := y.Map()
		So(err, ShouldBeNil)

		return data
	}
	ToYAML := func(tree map[interface{}]interface{}) string {
		y, err := yaml.Marshal(tree)
		So(err, ShouldBeNil)
		return string(y)
	}
	ReYAML := func(s string) string {
		return ToYAML(YAML(s))
	}
	RunPhaseTests := func(phase OperatorPhase, src string) {
		var test, input, dataflow, output string
		var current *string
		testPat := regexp.MustCompile(`^##+\s+(.*)\s*$`)

		convey := func() {
			if test != "" {
				Convey(test, func() {
					ev := &Evaluator{Tree: YAML(input)}

					ops, err := ev.DataFlow(phase)
					So(err, ShouldBeNil)

					// map data flow ops into 'dataflow:' YAML list
					var flow []map[string]string
					for _, op := range ops {
						flow = append(flow, map[string]string{op.where.String(): op.src})
					}
					So(ToYAML(map[interface{}]interface{}{"dataflow": flow}),
						ShouldEqual, ReYAML(dataflow))

					err = ev.RunPhase(phase)
					So(err, ShouldBeNil)
					So(ToYAML(ev.Tree), ShouldEqual, ReYAML(output))
				})
			}
		}

		s := bufio.NewScanner(strings.NewReader(src))
		for s.Scan() {
			if testPat.MatchString(s.Text()) {
				m := testPat.FindStringSubmatch(s.Text())
				convey()
				test, input, dataflow, output = m[1], "", "", ""
				continue
			}

			if s.Text() == "---" {
				if input == "" {
					current = &input
				} else if dataflow == "" {
					current = &dataflow
				} else {
					current = &output
				}
				continue
			}

			if current != nil {
				*current = *current + s.Text() + "\n"
			}
		}
		convey()
	}

	/*
	   ##     ## ######## ########   ######   ########
	   ###   ### ##       ##     ## ##    ##  ##
	   #### #### ##       ##     ## ##        ##
	   ## ### ## ######   ########  ##   #### ######
	   ##     ## ##       ##   ##   ##    ##  ##
	   ##     ## ##       ##    ##  ##    ##  ##
	   ##     ## ######## ##     ##  ######   ########
	*/

	Convey("Merge Phase", t, func() {
		RunPhaseTests(MergePhase, `
##################################################   can handle simplest case
---
templates:
  www:
    HA: enabled
    DR: disabled
host:
  web1:
    type: www
    <<<: (( inject templates.www ))

---
dataflow:
- host.web1.<<<: (( inject templates.www ))

---
templates:
  www:
    HA: enabled
    DR: disabled
host:
  web1:
    type: www
    HA: enabled
    DR: disabled


################################################   ignores EvalPhase operators
---
meta:
  template:
    check: (( param "stuff" ))
    baz: (( grab meta.template.foo ))
    str: (( concat "foo" meta.template.baz ))
    foo: bar

thing:
  <<<: (( inject meta.template ))

---
dataflow:
- thing.<<<: (( inject meta.template ))

---
meta:
  template:
    check: (( param "stuff" ))
    baz: (( grab meta.template.foo ))
    str: (( concat "foo" meta.template.baz ))
    foo: bar

thing:
  check: (( param "stuff" ))
  baz: (( grab meta.template.foo ))
  str: (( concat "foo" meta.template.baz ))
  foo: bar


############################################   handles nested (( inject ... )) calls
---
meta:
  template1:
    foo: bar
    baz: (( grab meta.template1.foo ))
  template2:
    <<<: (( inject meta.template1 ))
    xyzzy: nothing happens

thing:
  <<<: (( inject meta.template2 ))

---
dataflow:
- meta.template2.<<<: (( inject meta.template1 ))
- thing.<<<:          (( inject meta.template2 ))

---
meta:
  template1:
    foo: bar
    baz: (( grab meta.template1.foo ))
  template2:
    foo: bar
    baz: (( grab meta.template1.foo ))
    xyzzy: nothing happens

thing:
  foo: bar
  baz: (( grab meta.template1.foo ))
  xyzzy: nothing happens

############################################   handles nested (( inject ... )) through array aliasing

---
a:
  foo: bar

jobs:
  - name: b
    <<<: (( inject a ))
    boz: baz

c:
  <<<: (( inject jobs.b ))
  xy: zzy

---
dataflow:
- jobs.b.<<<: (( inject a ))
- c.<<<: (( inject jobs.b ))

---
a:
  foo: bar

jobs:
  - name: b
    foo: bar
    boz: baz

c:
  name: b
  foo: bar
  boz: baz
  xy: zzy



#########################################################   handles Inject into a list
---
meta:
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

---
dataflow:
- jobs.api_z1.<<<: (( inject meta.jobs.api ))
- jobs.api_z2.<<<: (( inject meta.jobs.api ))
- jobs.worker_z3.<<<: (( inject meta.jobs.worker ))
- jobs.db_z3.<<<: (( inject meta.jobs.db ))

---
meta:
  jobs:
    api:
      template: api
    worker:
      template: worker
    db:
      template: database

jobs:
  - name: api_z1
    template: api
  - name: api_z2
    template: api

  - name: worker_z3
    template: worker

  - name: db_z3
    template: database


#################################   preserves call-site keys on conflict in an Inject scenario
---
meta:
  template:
    foo: FOO
    bar: BAR

example:
  <<<: (( inject meta.template ))
  foo: foooo

---
dataflow:
- example.<<<: (( inject meta.template ))

---
meta:
  template:
    foo: FOO
    bar: BAR

example:
  foo: foooo
  bar: BAR


#############################   merges sub-tress common to inject sites and injected values
---
meta:
  template:
    properties:
      foo: bar
      baz: NOT-OVERRIDDEN

thing:
  <<<: (( inject meta.template ))
  properties:
    bar: baz
    baz: overridden

---
dataflow:
- thing.<<<: (( inject meta.template ))

---
meta:
  template:
    properties:
      foo: bar
      baz: NOT-OVERRIDDEN

thing:
  properties:
    foo: bar
    bar: baz
    baz: overridden

#################   merges name-indexed sub-arrays properly between call-site and inject site
---
meta:
  api_node:
    templates:
    - name: my_job
      release: my_release
    - name: my_other_job
      release: my_other_release
    properties:
      foo: bar
jobs:
  api_node:
    .: (( inject meta.api_node ))
    properties:
      this: that
    templates:
    - name: my_superspecial_job
      release: my_superspecial_release

---
dataflow:
- jobs.api_node..: (( inject meta.api_node ))

---
meta:
  api_node:
    templates:
    - name: my_job
      release: my_release
    - name: my_other_job
      release: my_other_release
    properties:
      foo: bar
jobs:
  api_node:
    properties:
      foo: bar
      this: that
    templates:
    - name: my_job
      release: my_release
    - name: my_other_job
      release: my_other_release
    - name: my_superspecial_job
      release: my_superspecial_release


#################   uses deep-copy semantics to handle overrides correctly on template re-use
---
meta:
  template:
    properties:
      key: DEFAULT
      sub:
        key: DEFAULT
foo:
  <<<: (( inject meta.template ))
  properties:
    key: FOO
    sub:
      key: FOO
bar:
  <<<: (( inject meta.template ))
  properties:
    key: BAR
    sub:
      key: BAR
boz:
  <<<: (( inject meta.template ))
  properties:
    key: BOZ
    sub:
      key: BOZ

---
dataflow:
- bar.<<<: (( inject meta.template ))
- boz.<<<: (( inject meta.template ))
- foo.<<<: (( inject meta.template ))

---
meta:
  template:
    properties:
      key: DEFAULT
      sub:
        key: DEFAULT
foo:
  properties:
    key: FOO
    sub:
      key: FOO
bar:
  properties:
    key: BAR
    sub:
      key: BAR
boz:
  properties:
    key: BOZ
    sub:
      key: BOZ


#########################################################  appends to injected arrays
---
meta:
  job:
    templates:
      - first
      - second
foo:
  <<<: (( inject meta.job ))
  templates:
    - third

---
dataflow:
- foo.<<<: (( inject meta.job ))

---
meta:
  job:
    templates:
      - first
      - second
foo:
  templates:
    - first
    - second
    - third
`)
	})

	Convey("Merge Phase Error Detection", t, func() {
		Convey("detects direct (a -> b -> a) cycles in data flow graph", func() {
			ev := &Evaluator{
				Tree: YAML(`
meta:
  bar:
    <<<: (( inject meta.foo ))
  foo:
    <<<: (( inject meta.bar ))
`),
			}

			_, err := ev.DataFlow(MergePhase)
			So(err, ShouldNotBeNil)
		})
	})

	/*
	   ######## ##     ##    ###    ##
	   ##       ##     ##   ## ##   ##
	   ##       ##     ##  ##   ##  ##
	   ######   ##     ## ##     ## ##
	   ##        ##   ##  ######### ##
	   ##         ## ##   ##     ## ##
	   ########    ###    ##     ## ########
	*/

	Convey("Eval Phase", t, func() {
		RunPhaseTests(EvalPhase, `
#############################################################   handles simple expressions
---
foo: (( concat "foo" ":" "bar" ))

---
dataflow:
- foo: (( concat "foo" ":" "bar" ))

---
foo: foo:bar


####################################################   handles simple reference expressions
---
meta:
  domain: foo.bar
domain: (( grab meta.domain ))

---
dataflow:
- domain: (( grab meta.domain ))

---
meta:
  domain: foo.bar
domain: foo.bar


#########################################   handles simple reference-or-literal expressions
---
meta:
  env: prod
domain:    (( grab meta.domain || "default-domain" ))
env:       (( grab meta.env || "sandbox" ))
instances: (( grab meta.size || 42 ))
nice:      (( grab meta.nice || -5 ))
pi:        (( grab math.CONSTANTS.pi || 3.14159 ))
delta:     (( grab meta.delta || .001 ))
secure:    (( grab meta.secure || true ))
online:    (( grab meta.online || false ))

---
dataflow:
- delta:     (( grab meta.delta || .001 ))
- domain:    (( grab meta.domain || "default-domain" ))
- env:       (( grab meta.env || "sandbox" ))
- instances: (( grab meta.size || 42 ))
- nice:      (( grab meta.nice || -5 ))
- online:    (( grab meta.online || false ))
- pi:        (( grab math.CONSTANTS.pi || 3.14159 ))
- secure:    (( grab meta.secure || true ))

---
meta:
  env: prod
domain: default-domain
env: prod
instances: 42
nice: -5
pi: 3.14159
delta: 0.001
secure: true
online: false


#####################################   handles true, TRUE, and True as boolean keywords
---
TrUe: sure
things:
- (( grab meta.enoent || true ))
- (( grab meta.enoent || TRUE ))
- (( grab meta.enoent || True ))
- (( grab meta.enoent || TrUe ))

---
dataflow:
- things.0: (( grab meta.enoent || true ))
- things.1: (( grab meta.enoent || TRUE ))
- things.2: (( grab meta.enoent || True ))
- things.3: (( grab meta.enoent || TrUe ))

---
TrUe: sure
things:
- true
- true
- true
- sure


#####################################   handles false, FALSE, and False as boolean keywords
---
FaLSe: why not?
things:
- (( grab meta.enoent || false ))
- (( grab meta.enoent || FALSE ))
- (( grab meta.enoent || False ))
- (( grab meta.enoent || FaLSe ))

---
dataflow:
- things.0: (( grab meta.enoent || false ))
- things.1: (( grab meta.enoent || FALSE ))
- things.2: (( grab meta.enoent || False ))
- things.3: (( grab meta.enoent || FaLSe ))

---
FaLSe: why not?
things:
- false
- false
- false
- why not?


######################   handles ~, nil, Nil, NIL, null, Null, and NULL as the nil keyword
---
NuLL: 4 (haha ruby joke)
things:
- (( grab meta.enoent || ~ ))
- (( grab meta.enoent || nil ))
- (( grab meta.enoent || Nil ))
- (( grab meta.enoent || NIL ))
- (( grab meta.enoent || null ))
- (( grab meta.enoent || Null ))
- (( grab meta.enoent || NULL ))
- (( grab meta.enoent || NuLL ))

---
dataflow:
- things.0: (( grab meta.enoent || ~ ))
- things.1: (( grab meta.enoent || nil ))
- things.2: (( grab meta.enoent || Nil ))
- things.3: (( grab meta.enoent || NIL ))
- things.4: (( grab meta.enoent || null ))
- things.5: (( grab meta.enoent || Null ))
- things.6: (( grab meta.enoent || NULL ))
- things.7: (( grab meta.enoent || NuLL ))

---
NuLL: 4 (haha ruby joke)
things:
- null
- null
- null
- null
- null
- null
- null
- 4 (haha ruby joke)




#########################################   handles simple reference-or-nil expressions
---
domain: (( grab meta.domain || nil ))
env:    (( grab meta.env || ~ ))
site:   (( grab meta.site || null ))

---
dataflow:
- domain: (( grab meta.domain || nil ))
- env:    (( grab meta.env || ~ ))
- site:   (( grab meta.site || null ))

---
domain: ~
env: ~
site: ~


##################################   stops at the first concrete (possibly false) expression
---
meta:
  other: FAIL

#
# should stop here ----------.   (because it's resolvable, even if it
#                            |    evaluates to a traditionally non-true value)
#                            v
foo: (( grab meta.enoent || false || meta.other || "failed" ))

---
dataflow:
- foo: (( grab meta.enoent || false || meta.other || "failed" ))

---
meta:
  other: FAIL
foo: false


##################################   stops at the first concrete (possibly 0) expression
---
meta:
  other: FAIL

#
# should stop here ----------.   (because it's resolvable, even if it
#                            |    evaluates to a traditionally non-true value)
#                            v
foo: (( grab meta.enoent || 0 || meta.other || "failed" ))

---
dataflow:
- foo: (( grab meta.enoent || 0 || meta.other || "failed" ))

---
meta:
  other: FAIL
foo: 0


##################################   stops at the first concrete (possibly nil) expression
---
meta:
  other: FAIL

#
# should stop here ----------.   (because it's resolvable, even if it
#                            |    evaluates to a traditionally non-true value)
#                            v
foo: (( grab meta.enoent || nil || meta.other || "failed" ))

---
dataflow:
- foo: (( grab meta.enoent || nil || meta.other || "failed" ))

---
meta:
  other: FAIL
foo: ~


###############################################  handles concrete expression in the middle
---
meta:
  second: SECOND
foo: (( grab meta.first || meta.second || "unspecified" ))

---
dataflow:
- foo: (( grab meta.first || meta.second || "unspecified" ))

---
meta:
  second: SECOND
foo: SECOND


#################################   skips short-circuited alternates in Data Flow analysis
---
meta:
  foo: FOO
  bar: (( grab meta.foo ))
  boz: (( grab meta.foo ))

foo: (( grab meta.bar || "foo?" || meta.boz ))
# NOTE: meta.boz in $.foo is exempt from DFA, because the "foo?" literal
#       will *always* stop evaluation of the expression

bar: (( grab meta.xyzzy || "bar?" || meta.boz ))
# NOTE: same with $.bar; meta.boz is not in play

---
dataflow:
- bar: (( grab meta.xyzzy || "bar?" || meta.boz ))
- meta.bar: (( grab meta.foo ))
- meta.boz: (( grab meta.foo ))
- foo: (( grab meta.bar || "foo?" || meta.boz ))

---
meta:
  foo: FOO
  bar: FOO
  boz: FOO
foo: FOO
bar: bar?


#####################################   handles Data Flow dependencies for all expressions
---
meta:
  domain: example.com
  web: (( concat "www.", meta.domain || "sandbox.example.com" ))
api:
  endpoint: (( grab meta.web || meta.domain || ~ ))

---
dataflow:
- meta.web: (( concat "www.", meta.domain || "sandbox.example.com" ))
- api.endpoint: (( grab meta.web || meta.domain || ~ ))

---
meta:
  domain: example.com
  web: www.example.com
api:
  endpoint: www.example.com


###############################  handles indirect addressing of lists-of-maps in Data Flow
---
a:  (( grab b.squad.value ))
b:
  - name: squad
    value: (( grab c.value ))
c:
  value: VALUE
d: (( grab b.squad.value ))
e: (( grab c.value ))

---
dataflow:
- b.squad.value: (( grab c.value ))
- e:         (( grab c.value ))
- a:         (( grab b.squad.value ))
- d:         (( grab b.squad.value ))

---
a: VALUE
b:
  - name: squad
    value: VALUE
c:
  value: VALUE
d: VALUE
e: VALUE


#################################   handles multiple space-separated or-operands properly
---
meta:
  foo: FOO
  bar: BAR
foobar: (( concat meta.foo || "foo"  meta.bar || "bar" ))
fooboz: (( concat meta.foo || "foo"  meta.boz || "boz" ))

---
dataflow:
- foobar: (( concat meta.foo || "foo"  meta.bar || "bar" ))
- fooboz: (( concat meta.foo || "foo"  meta.boz || "boz" ))

---
meta:
  foo: FOO
  bar: BAR
foobar: FOOBAR
fooboz: FOOboz


#################################   handles multiple comma-seaprated or-operands properly
---
meta:
  foo: FOO
  bar: BAR
foobar: (( concat meta.foo || "foo", meta.bar || "bar" ))
fooboz: (( concat meta.foo || "foo", meta.boz || "boz" ))

---
dataflow:
- foobar: (( concat meta.foo || "foo", meta.bar || "bar" ))
- fooboz: (( concat meta.foo || "foo", meta.boz || "boz" ))

---
meta:
  foo: FOO
  bar: BAR
foobar: FOOBAR
fooboz: FOOboz


############################################   can handle simple map-based Replace actions
---
meta:
  domain: sandbox.example.com
  web: (( grab meta.domain ))
urls:
  home: (( concat "http://www." meta.web ))

---
dataflow:
- meta.web: (( grab meta.domain ))
- urls.home: (( concat "http://www." meta.web ))

---
meta:
  domain: sandbox.example.com
  web: sandbox.example.com
urls:
  home: http://www.sandbox.example.com


############################   can handle Replacement actions where the new value is a list
---
meta:
  things:
    - one
    - two
grocery:
  list: (( grab meta.things ))

---
dataflow:
- grocery.list: (( grab meta.things ))

---
meta:
  things:
    - one
    - two
grocery:
  list:
    - one
    - two


############################   can handle Replacement actions where the call site is a list
---
meta:
  first:  2nd
  second: 1st
sorted:
  list:
    - (( grab meta.second ))
    - (( grab meta.first ))

---
dataflow:
- sorted.list.0: (( grab meta.second ))
- sorted.list.1: (( grab meta.first ))

---
meta:
  first:  2nd
  second: 1st
sorted:
  list:
    - 1st
    - 2nd


###################   can handle Replacement actions where the call site is inside of a list
---
meta:
  prod: production
  sandbox: sb322
boxen:
  - name: www
    env: (( grab meta.prod ))
  - name: wwwtest
    env: (( grab meta.sandbox ))

---
dataflow:
- boxen.www.env: (( grab meta.prod ))
- boxen.wwwtest.env: (( grab meta.sandbox ))

---
meta:
  prod: production
  sandbox: sb322
boxen:
  - name: www
    env: production
  - name: wwwtest
    env: sb322


########################################   handles sequentially dependent (( grab ... )) calls
---
meta:
  foo:  FOO
  bar:  (( grab meta.foo ))
  baz:  (( grab meta.bar ))
  quux: (( grab meta.baz ))
  boz:  (( grab meta.quux ))

---
dataflow:
- meta.bar:   (( grab meta.foo ))
- meta.baz:   (( grab meta.bar ))
- meta.quux:  (( grab meta.baz ))
- meta.boz:   (( grab meta.quux ))

---
meta:
  foo:  FOO
  bar:  FOO
  baz:  FOO
  quux: FOO
  boz:  FOO


################################   handles sequentially dependent calls, regardless of operator
---
meta:
  foo: FOO
  bar: (( grab meta.foo ))
  baz: (( grab meta.bar ))
  quux: (( concat "literal:" meta.baz ))
  boz: (( grab meta.quux ))

---
dataflow:
- meta.bar:  (( grab meta.foo ))
- meta.baz:  (( grab meta.bar ))
- meta.quux: (( concat "literal:" meta.baz ))
- meta.boz:  (( grab meta.quux ))

---
meta:
  foo: FOO
  bar: FOO
  baz: FOO
  quux: literal:FOO
  boz:  literal:FOO


##############################################   handles operators dependencies inside of lists
---
meta:
  - FOO
  - (( grab meta.0 ))
  - (( grab meta.1 ))
  - (( concat "literal:" meta.2 ))
  - (( grab meta.3 ))

---
dataflow:
- meta.1:  (( grab meta.0 ))
- meta.2:  (( grab meta.1 ))
- meta.3:  (( concat "literal:" meta.2 ))
- meta.4:  (( grab meta.3 ))

---
meta:
  - FOO
  - FOO
  - FOO
  - literal:FOO
  - literal:FOO


#################################################   handles deep copy in data flow graph
---
meta:
  first: [ a, b, c ]
  second: (( grab meta.first ))
  third:  (( grab meta.second ))
  gotcha: (( grab meta.third.0 ))

---
dataflow:
- meta.second: (( grab meta.first ))
- meta.third:  (( grab meta.second ))
- meta.gotcha: (( grab meta.third.0 ))

# (the key point here is that meta.third.0 doesn't exist in the tree until we start
# evaluating, but we still need to get the order correct; we should have a dep on
# meta.third, and hope that run-time resolution puts an array there for us to find...)

---
meta:
  first:  [ a, b, c ]
  second: [ a, b, c ]
  third:  [ a, b, c ]
  gotcha:   a


###############################  handles implicit static_ip dependency on networks.*.name
---
meta:
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
      - name: real-prod # must be literal to avoid non-determinism
        static_ips: (( static_ips 1 2 3 4 ))

---
dataflow:
- networks.0.name:                         (( concat meta.net "-prod" ))
- jobs.job1.networks.real-prod.static_ips: (( static_ips 1 2 3 4 ))

---
meta:
  net: real
  environment: prod
  size: 4
networks:
  - name: real-prod
    subnets:
    - static: [ 10.0.0.5 - 10.0.0.100 ]
jobs:
  - name: job1
    instances: 4
    networks:
      - name: real-prod
        static_ips:
            # skip IP[0]!
          - 10.0.0.6
          - 10.0.0.7
          - 10.0.0.8
          - 10.0.0.9


####################   handles implicit static_ip dependency on jobs.*.networks.*.name
---
meta:
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

---
dataflow:
- jobs.job1.networks.0.name:        (( grab meta.environment ))
- jobs.job1.networks.0.static_ips:  (( static_ips 1 2 3 4 ))

---
meta:
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
      - name: prod
        static_ips:
            # skip IP[0]!
          - 10.0.0.6
          - 10.0.0.7
          - 10.0.0.8
          - 10.0.0.9


##########################   handles (( static_ips ... )) call and a subsequent (( grab ... ))
---
jobs:
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

---
dataflow:
- jobs.api_z1.networks.net1.static_ips: (( static_ips(0, 1, 2) ))
- properties.api_servers:       (( grab jobs.api_z1.networks.net1.static_ips ))

---
jobs:
- name: api_z1
  instances: 1
  networks:
  - name: net1
    static_ips:
    - 192.168.1.2

networks:
- name: net1
  subnets:
    - static: [192.168.1.2 - 192.168.1.30]

properties:
  api_servers:
    - 192.168.1.2


#############################################  handles static_ip across multiple static ranges
---
jobs:
- name: api_z1
  instances: 4
  networks:
  - name: net1
    static_ips: (( static_ips 0 1 2 3 ))

networks:
- name: net1
  subnets:
    - static:
      - 10.0.0.2 - 10.0.0.3      #  2 ips
      - 10.0.0.90                # +1
      - 10.0.0.100 - 10.0.0.103  # +4

---
dataflow:
- jobs.api_z1.networks.net1.static_ips: (( static_ips 0 1 2 3 ))

---
jobs:
- name: api_z1
  instances: 4
  networks:
  - name: net1
    static_ips:
    - 10.0.0.2
    - 10.0.0.3
    - 10.0.0.90
    - 10.0.0.100

networks:
- name: net1
  subnets:
    - static:
      - 10.0.0.2 - 10.0.0.3      #  2 ips
      - 10.0.0.90                # +1
      - 10.0.0.100 - 10.0.0.103  # +4


#########################################  Basic test of (( cartesian-product .... )) operator
---
meta:
  hosts:
    - a.example.com
    - b.example.com
    - c.example.com
  port: 8088

hosts: (( cartesian-product meta.hosts ":" meta.port ))

---
dataflow:
- hosts: (( cartesian-product meta.hosts ":" meta.port ))

---
meta:
  hosts:
    - a.example.com
    - b.example.com
    - c.example.com
  port: 8088

hosts:
  - a.example.com:8088
  - b.example.com:8088
  - c.example.com:8088


####################################  (( cartesian-product .... )) of an empty component array
---
meta:
  hosts: []
  port: 8088

hosts: (( cartesian-product meta.hosts ":" meta.port ))
ports: (( cartesian-product meta.port  ":" meta.hosts ))

---
dataflow:
- hosts: (( cartesian-product meta.hosts ":" meta.port ))
- ports: (( cartesian-product meta.port  ":" meta.hosts ))

---
meta:
  hosts: []
  port: 8088

hosts: []
ports: []


##########################################################  1-ary (( cartesian-product .... ))
---
meta:
  - [a, b, c]

all: (( cartesian-product meta[0] ))

---
dataflow:
- all: (( cartesian-product meta[0] ))

---
meta:
  - [a, b, c]
all: [a, b, c]


##########################################################  n-ary (( cartesian-product .... ))
---
meta:
  - [a, b, c]
  - [1, 2, 3]
  - [x, 'y', z]

all: (( cartesian-product meta[0] meta[1] meta[2] ))

---
dataflow:
- all: (( cartesian-product meta[0] meta[1] meta[2] ))

---
meta:
  - [a, b, c]
  - [1, 2, 3]
  - [x, 'y', z]
all:
  - a1x
  - a1y
  - a1z
  - a2x
  - a2y
  - a2z
  - a3x
  - a3y
  - a3z
  - b1x
  - b1y
  - b1z
  - b2x
  - b2y
  - b2z
  - b3x
  - b3y
  - b3z
  - c1x
  - c1y
  - c1z
  - c2x
  - c2y
  - c2z
  - c3x
  - c3y
  - c3z


###########################################  (( cartesian-product ... )) with grab'd arguments
---
meta:
  first: [a, b]
  second: [1, 2]
  third:  (( grab meta.second ))

all: (( cartesian-product meta.first "," meta.third ))

---
dataflow:
- meta.third: (( grab meta.second ))
- all: (( cartesian-product meta.first "," meta.third ))

---
meta:
  first:  [a, b]
  second: [1, 2]
  third:  [1, 2]

all:
  - a,1
  - a,2
  - b,1
  - b,2

###################################  (( cartesian-product ... )) with grab'd sublist arguments
---
meta:
  first: [a, b]
  second:
    - x
    - (( grab meta.first[0] ))

all: (( cartesian-product meta.first "," meta.second ))

---
dataflow:
- meta.second.1: (( grab meta.first[0] ))
- all: (( cartesian-product meta.first "," meta.second ))

---
meta:
  first:  [a, b]
  second: [x, a]

all:
  - a,x
  - a,a
  - b,x
  - b,a


###########################################  can extract keys via the (( keys ... )) operator
---
meta:
  config:
    first: this is the first value
    second:
      value: the second
keys: (( keys meta.config ))

---
dataflow:
- keys: (( keys meta.config ))

---
meta:
  config:
    first: this is the first value
    second:
      value: the second
keys:
  - first
  - second


###########################################  can extract keys from multiple maps
---
meta:
  config:
    first: this is the first value
    second:
      value: the second
  alt:
    third: third config
keys: (( keys meta.config meta.alt ))

---
dataflow:
- keys: (( keys meta.config meta.alt ))

---
meta:
  config:
    first: this is the first value
    second:
      value: the second
  alt:
    third: third config
keys:
  - first
  - second
  - third

#################################### (( join ... )) an array with (( grab ...))
---
greeting: hello

z:
- (( grab greeting ))
- world

output: (( join " " z ))
---
dataflow:
- z.0: (( grab greeting ))
- output: (( join " " z ))
---
greeting: hello
output: hello world
z:
- hello
- world
#################################### (( join ... )) an array with several (( grab ...))s
---
greeting: hello
greeting2: world

z:
- (( grab greeting ))
- (( grab greeting2 ))

output: (( join " " z ))
---
dataflow:
- z.0: (( grab greeting ))
- z.1: (( grab greeting2 ))
- output: (( join " " z ))
---
greeting: hello
greeting2: world
output: hello world
z:
- hello
- world
#################################### (( join ... )) a string reference with a grab
---
greeting: hello
z_one: (( grab greeting ))
z_two: world
output:
- (( join " " z_one z_two ))
---
dataflow:
- z_one: (( grab greeting ))
- output.0: (( join " " z_one z_two ))
---
greeting: hello
output:
- hello world
z_one: hello
z_two: world


################################################   basic escape sequence handling
---
test: ""
cr:   (( concat test "a\rb" ))
nl:   (( concat test "a\nb" ))
tab:  (( concat test "a\tb" ))
back: (( concat test "a\\b" ))
dq:   (( concat test "a\"b" ))
sq:   (( concat test "a\'b" ))

---
dataflow:
- back: (( concat test "a\\b" ))
- cr:   (( concat test "a\rb" ))
- dq:   (( concat test "a\"b" ))
- nl:   (( concat test "a\nb" ))
- sq:   (( concat test "a\'b" ))
- tab:  (( concat test "a\tb" ))

---
test: ""
cr:   "a\rb"
nl:   "a\nb"
tab:  "a\tb"
back: a\b
dq:   'a"b'
sq:   "a'b"


#############################################   repeated escape sequence handling
---
compound: (( concat "Line1\nLine2\nLine3" "\n" "Line4\ttabbed\n" ))

---
dataflow:
- compound: (( concat "Line1\nLine2\nLine3" "\n" "Line4\ttabbed\n" ))

---
compound: |
  Line1
  Line2
  Line3
  Line4	tabbed


########################################   concat certs with newlines (escape seq)
---
cert: |-
  -- BEGIN CERT --
  unei3Eet2mahbou8
  weiXi7choo7ufei8
  --- END CERT ---

key: |-
  -- BEGIN KEY ---
  chaev0Gai3Baedul
  noithaifu0ree0Ka
  shoowuBaoti4chee
  -- END KEY -----

combined: (( concat cert "\n" key "\n" ))

---
dataflow:
- combined: (( concat cert "\n" key "\n" ))

---
cert: |-
  -- BEGIN CERT --
  unei3Eet2mahbou8
  weiXi7choo7ufei8
  --- END CERT ---

key: |-
  -- BEGIN KEY ---
  chaev0Gai3Baedul
  noithaifu0ree0Ka
  shoowuBaoti4chee
  -- END KEY -----

combined: |
  -- BEGIN CERT --
  unei3Eet2mahbou8
  weiXi7choo7ufei8
  --- END CERT ---
  -- BEGIN KEY ---
  chaev0Gai3Baedul
  noithaifu0ree0Ka
  shoowuBaoti4chee
  -- END KEY -----

`)
	})

	Convey("Eval Phase Error Detection", t, func() {
		Convey("detects direct (a -> b -> a) cycles in data flow graph", func() {
			ev := &Evaluator{
				Tree: YAML(`
meta:
  bar: (( grab meta.foo ))
  foo: (( grab meta.bar ))
`),
			}

			_, err := ev.DataFlow(EvalPhase)
			So(err, ShouldNotBeNil)
		})

		Convey("detects indirect (a -> b -> c -> a) cycles in data flow graph", func() {
			ev := &Evaluator{
				Tree: YAML(`
meta:
  foo: (( grab meta.bar ))
  bar: (( grab meta.baz ))
  baz: (( grab meta.foo ))
`),
			}

			_, err := ev.DataFlow(EvalPhase)
			So(err, ShouldNotBeNil)
		})

		Convey("detects indirect cycles created through operand data flow", func() {
			ev := &Evaluator{
				Tree: YAML(`
meta:
  foo: (( grab meta.bar ))
  bar: (( grab meta.baz ))
  baz: (( grab meta.enoent || meta.foo ))
`),
			}

			_, err := ev.DataFlow(EvalPhase)
			So(err, ShouldNotBeNil)
		})

		Convey("detects allocation conflicts of static IP addresses", func() {
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

			err := ev.RunPhase(EvalPhase)
			So(err, ShouldNotBeNil)
		})

		Convey("detects unsatisfied (( param )) inside of a (( grab ... )) call", func() {
			ev := &Evaluator{
				Tree: YAML(`
meta:
  key: (( param "you must specify this" ))
value: (( grab meta.key ))
`),
			}

			err := ev.RunPhase(ParamPhase)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "1 error(s) detected")
			So(err.Error(), ShouldContainSubstring, "you must specify this")

			err = ev.RunPhase(EvalPhase)
			So(err, ShouldBeNil)
		})

		Convey("detects unsatisfied (( param )) inside of a (( concat ... )) call", func() {
			ev := &Evaluator{
				Tree: YAML(`
---
meta:
  key: (( param "you must specify this" ))
value: (( concat "key=" meta.key ))
`),
			}

			err := ev.RunPhase(ParamPhase)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "1 error(s) detected")
			So(err.Error(), ShouldContainSubstring, "you must specify this")

			err = ev.RunPhase(EvalPhase)
			So(err, ShouldBeNil)
		})

		Convey("handles non-list (direct) args to (( cartesian-product ... ))", func() {
			ev := &Evaluator{
				Tree: YAML(`
meta:
  list: [a,b,c]
all: (( cartesian-product meta meta.list ))
`),
			}

			err := ev.RunPhase(EvalPhase)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "cartesian-product operator only accepts arrays and string values")
		})

		Convey("treats list-of-lists args to (( cartesian-product ... )) as an error", func() {
			ev := &Evaluator{
				Tree: YAML(`
meta:
  list:
    - [a,b,c]
    - [d,e,f]
    - [g]
all: (( cartesian-product meta.list meta.list ))
`),
			}

			err := ev.RunPhase(EvalPhase)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "cartesian-product operator can only operate on lists of scalar values")
		})

		Convey("treats list-of-maps args to (( cartesian-product ... )) as an error", func() {
			ev := &Evaluator{
				Tree: YAML(`
meta:
  list:
    - name: a
    - name: b
    - name: c
all: (( cartesian-product meta.list meta.list ))
`),
			}

			err := ev.RunPhase(EvalPhase)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "cartesian-product operator can only operate on lists of scalar values")
		})

		Convey("(( cartesian-product ... )) requires an argument", func() {
			ev := &Evaluator{
				Tree: YAML(`
all: (( cartesian-product ))
`),
			}

			err := ev.RunPhase(EvalPhase)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no arguments specified to (( cartesian-product ... ))")
		})

		Convey("(( keys ... )) requires an argument", func() {
			ev := &Evaluator{
				Tree: YAML(`
keys: (( keys ))
`),
			}

			err := ev.RunPhase(EvalPhase)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "no arguments specified to (( keys ... ))")
		})

		Convey("treats attempt to call (( keys ... )) on a literal as an error", func() {
			ev := &Evaluator{
				Tree: YAML(`
meta:
  test: is this a map?
keys: (( keys meta.test ))
`),
			}

			err := ev.RunPhase(EvalPhase)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "$.keys: meta.test is not a map")
		})

		Convey("treats attempt to call (( keys ... )) on a list as an error", func() {
			ev := &Evaluator{
				Tree: YAML(`
meta:
  test:
    - but wait
    - this is not
    - a map...
keys: (( keys meta.test ))
`),
			}

			err := ev.RunPhase(EvalPhase)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "$.keys: meta.test is not a map")
		})
	})
}
