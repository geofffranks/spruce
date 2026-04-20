package main_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// runSpruce is a helper to run spruce with args and return a session
func runSpruce(args ...string) *gexec.Session {
	cmd := exec.Command(sprucePath, args...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

// runSpruceWithStdin runs spruce with stdin and extra args
func runSpruceWithStdin(stdin string, args ...string) *gexec.Session {
	cmd := exec.Command(sprucePath, args...)
	cmd.Stdin = strings.NewReader(stdin)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

// runSpruceWithEnv runs spruce with extra env vars and args
func runSpruceWithEnv(env []string, args ...string) *gexec.Session {
	cmd := exec.Command(sprucePath, args...)
	cmd.Env = append(os.Environ(), env...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

var _ = Describe("main()", func() {
	It("Should output usage if bad args are passed", func() {
		session := runSpruce("fdsafdada")
		Eventually(session, "10s").Should(gexec.Exit(1))
	})

	It("Should output usage if no args at all", func() {
		session := runSpruce()
		Eventually(session, "10s").Should(gexec.Exit(1))
	})

	It("Should error if no args to merge and no files listed", func() {
		session := runSpruce("merge")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).To(ContainSubstring("Error reading STDIN: no data found. Did you forget to pipe data to STDIN, or specify yaml files to merge?"))
	})

	Context("Should output version", func() {
		It("When '-v' is specified", func() {
			session := runSpruce("-v")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("spruce - Version"))
			Expect(string(session.Err.Contents())).To(BeEmpty())
		})

		It("When '--version' is specified", func() {
			session := runSpruce("--version")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring("spruce - Version"))
			Expect(string(session.Err.Contents())).To(BeEmpty())
		})
	})

	It("Should panic on errors merging docs", func() {
		session := runSpruce("merge", "../../assets/merge/bad.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).To(ContainSubstring("../../assets/merge/bad.yml: Root of YAML document is not a hash/map:"))
	})

	It("Should output merged yaml on success", func() {
		session := runSpruce("merge", "../../assets/merge/first.yml", "../../assets/merge/second.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`array_append:
- one
- two
- three
array_default:
- FIRST
- SECOND
- third
array_inline:
- name: first_elem
  val: overwritten
- second_elem was overwritten
- third elem is appended
array_map_default:
- k1: key 1
  k2: updated
  name: AAA
- k2: final
  k3: original
  name: BBB
array_prepend:
- three
- four
- five
array_replace:
- - 1
  - 2
  - 3
key: overridden
map:
  key: value
  key2: val2

`))
	})

	It("Should output merged yaml with multi-doc enabled", func() {
		session := runSpruce("merge", "-m", "../../assets/merge/multi-doc.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`doc:
  data:
    test01: stuff
    test02: morestuff

`))
	})

	It("Should not evaluate spruce logic when --no-eval", func() {
		session := runSpruce("merge", "--skip-eval", "../../assets/no-eval/first.yml", "../../assets/no-eval/second.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`injected_jobs:
  .: (( inject jobs ))
jobs:
- name: consul
- name: route
- name: cell
- name: cc_bridge
param: (( param "Fill this in later" ))
properties:
  loggregator: true
  no_eval: (( grab property ))
  no_prune: (( prune ))
  not_empty: not_empty

`))
	})

	It("Should execute --prunes when --no-eval", func() {
		session := runSpruce("merge", "--skip-eval", "--prune", "jobs", "../../assets/no-eval/first.yml", "../../assets/no-eval/second.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`injected_jobs:
  .: (( inject jobs ))
param: (( param "Fill this in later" ))
properties:
  loggregator: true
  no_eval: (( grab property ))
  no_prune: (( prune ))
  not_empty: not_empty

`))
	})

	It("Should execute --cherry-picks when --no-eval", func() {
		session := runSpruce("merge", "--skip-eval", "--cherry-pick", "properties", "../../assets/no-eval/first.yml", "../../assets/no-eval/second.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`properties:
  loggregator: true
  no_eval: (( grab property ))
  no_prune: (( prune ))
  not_empty: not_empty

`))
	})

	It("Should handle de-referencing", func() {
		session := runSpruce("merge", "../../assets/dereference/first.yml", "../../assets/dereference/second.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`jobs:
- name: my-server
  static_ips:
  - 192.168.1.0
properties:
  client:
    servers:
    - 192.168.1.0

`))
	})

	It("De-referencing cyclical datastructures should throw an error", func() {
		session := runSpruce("merge", "../../assets/dereference/cyclic-data.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Out.Contents())).To(BeEmpty())
		Expect(string(session.Err.Contents())).To(ContainSubstring("max recursion depth. You seem to have a self-referencing dataset\n"))
	})

	It("Dereferencing multiple values should behave as desired", func() {
		session := runSpruce("merge", "../../assets/dereference/multi-value.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`jobs:
- instances: 1
  name: api_z1
  networks:
  - name: net1
    static_ips:
    - 192.168.1.2
- instances: 1
  name: api_z2
  networks:
  - name: net2
    static_ips:
    - 192.168.2.2
networks:
- name: net1
  subnets:
  - cloud_properties: random
    static:
    - 192.168.1.2 - 192.168.1.30
- name: net2
  subnets:
  - cloud_properties: random
    static:
    - 192.168.2.2 - 192.168.2.30
properties:
  api_server_primary: 192.168.1.2
  api_servers:
  - 192.168.1.2
  - 192.168.2.2

`))
	})

	It("Should output error on bad de-reference", func() {
		session := runSpruce("merge", "../../assets/dereference/bad.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).To(ContainSubstring("$.bad.dereference: unable to resolve `my.value`"))
	})

	It("Pruning should happen after de-referencing", func() {
		session := runSpruce("merge", "--prune", "jobs", "--prune", "properties.client.servers", "../../assets/dereference/first.yml", "../../assets/dereference/second.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`properties:
  client: {}

`))
	})

	It("can dereference ~ / null values", func() {
		session := runSpruce("merge", "--prune", "meta", "../../assets/dereference/null.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`value: null

`))
	})

	It("can dereference nestedly", func() {
		session := runSpruce("merge", "../../assets/dereference/multi.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`name1: name
name2: name
name3: name
name4: name

`))
	})

	It("static_ips() failures return errors to the user", func() {
		session := runSpruce("merge", "../../assets/static_ips/jobs.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).To(ContainSubstring(".static_ips: `$.networks` could not be found in the datastructure\n"))
		Expect(string(session.Out.Contents())).To(BeEmpty())
	})

	It("static_ips() get resolved, and are resolved prior to dereferencing", func() {
		session := runSpruce("merge", "../../assets/static_ips/properties.yml", "../../assets/static_ips/jobs.yml", "../../assets/static_ips/network.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`jobs:
- instances: 3
  name: api_z1
  networks:
  - name: net1
    static_ips:
    - 10.0.0.2
    - 10.0.0.3
    - 10.0.0.4
networks:
- name: net1
  subnets:
  - static:
    - 10.0.0.2 - 10.0.0.20
properties:
  api_servers:
  - 10.0.0.2
  - 10.0.0.3
  - 10.0.0.4

`))
	})

	It("Included yaml file is escaped", func() {
		session := runSpruceWithEnv([]string{"SPRUCE_FILE_BASE_PATH=../../assets/file_operator"}, "merge", "../../assets/file_operator/test.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(`content:
  meta_test:
    stuff: |
      ---
      meta:
        filename: test.yml

      content:
        meta_test:
          stuff: (( file meta.filename ))
meta:
  filename: test.yml

`))
		Expect(string(session.Err.Contents())).To(BeEmpty())
	})

	It("Parameters override their requirement", func() {
		session := runSpruce("merge", "../../assets/params/global.yml", "../../assets/params/good.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(`cpu: 3
nested:
  key:
    override: true
networks:
- true
storage: 4096

`))
		Expect(string(session.Err.Contents())).To(BeEmpty())
	})

	It("Parameters must be specified", func() {
		session := runSpruce("merge", "../../assets/params/global.yml", "../../assets/params/fail.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Out.Contents())).To(BeEmpty())
		Expect(string(session.Err.Contents())).To(ContainSubstring("$.nested.key.override: provide nested override\n"))
	})

	It("Pruning takes place after parameters", func() {
		session := runSpruce("merge", "--prune", "nested", "../../assets/params/global.yml", "../../assets/params/fail.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).To(Equal(`1 error(s) detected:
 - $.nested.key.override: provide nested override


`))
		Expect(string(session.Out.Contents())).To(BeEmpty())
	})

	It("string concatenation works", func() {
		session := runSpruce("merge", "--prune", "local", "--prune", "env", "--prune", "cluster", "../../assets/concat/concat.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`ident: c=mjolnir/prod;1234567890-abcdef

`))
	})

	It("string concatenation handles non-strings correctly", func() {
		session := runSpruce("merge", "--prune", "local", "../../assets/concat/coerce.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`url: http://domain.example.com/?v=1.3&rev=42

`))
	})

	It("string concatenation failure detected", func() {
		session := runSpruce("merge", "../../assets/concat/fail.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Out.Contents())).To(BeEmpty())
		Expect(string(session.Err.Contents())).To(ContainSubstring("$.ident: unable to resolve `local.sites.42.uuid`:"))
	})

	It("string concatenation handles multiple levels of reference", func() {
		session := runSpruce("merge", "../../assets/concat/multi.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`bar: quux.bar
baz: quux.bar.baz
foo: quux.bar.baz.foo
quux: quux

`))
	})

	It("string concatenation handles infinite loop self-reference", func() {
		session := runSpruce("merge", "../../assets/concat/loop.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Out.Contents())).To(BeEmpty())
		Expect(string(session.Err.Contents())).To(ContainSubstring("cycle detected"))
	})

	It("only param errors are displayed, if present", func() {
		session := runSpruce("merge", "../../assets/errors/multi.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Out.Contents())).To(BeEmpty())
		Expect(string(session.Err.Contents())).To(Equal("" +
			"1 error(s) detected:\n" +
			" - $.an-error: missing param!\n" +
			"\n\n" +
			""))
	})

	It("multiple errors of the same type on the same level are displayed", func() {
		session := runSpruce("merge", "../../assets/errors/multi2.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Out.Contents())).To(BeEmpty())
		Expect(string(session.Err.Contents())).To(Equal("" +
			"3 error(s) detected:\n" +
			" - $.a: first\n" +
			" - $.b: second\n" +
			" - $.c: third\n" +
			"\n\n" +
			""))
	})

	It("json command converts YAML to JSON", func() {
		session := runSpruce("json", "../../assets/json/in.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`{"map":{"list":["string",42,{"map":"of things"}]}}` + "\n"))
	})

	It("json command handles malformed YAML", func() {
		session := runSpruce("json", "../../assets/json/malformed.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Out.Contents())).To(BeEmpty())
		Expect(string(session.Err.Contents())).To(ContainSubstring("Root of YAML document is not a hash/map:"))
	})

	It("vaultinfo lists vault calls in given file", func() {
		session := runSpruce("vaultinfo", "../../assets/vaultinfo/single.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(`secrets:
- key: secret/bar:beep
  references:
  - meta.foo

`))
		Expect(string(session.Err.Contents())).To(BeEmpty())
	})

	It("vaultinfo can handle multiple references to the same key", func() {
		session := runSpruce("vaultinfo", "../../assets/vaultinfo/duplicate.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(`secrets:
- key: secret/bar:beep
  references:
  - meta.foo
  - meta.otherfoo

`))
		Expect(string(session.Err.Contents())).To(BeEmpty())
	})

	It("vaultinfo can handle there being no vault references", func() {
		session := runSpruce("vaultinfo", "../../assets/vaultinfo/novault.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(`secrets: []

`))
		Expect(string(session.Err.Contents())).To(BeEmpty())
	})

	It("vaultinfo can handle concatenated vault secrets", func() {
		session := runSpruce("vaultinfo", "../../assets/vaultinfo/concat.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(`secrets:
- key: imaprefix/beep:boop
  references:
  - foo.bar
- key: imaprefix/cup:cake
  references:
  - foo.bat
- key: imaprefix/hello:world
  references:
  - foo.wom

`))
		Expect(string(session.Err.Contents())).To(BeEmpty())
	})

	It("vaultinfo can merge multiple files", func() {
		session := runSpruce("vaultinfo", "../../assets/vaultinfo/merge1.yml", "../../assets/vaultinfo/merge2.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(`secrets:
- key: secret/foo:bar
  references:
  - foo
- key: secret/meep:meep
  references:
  - bar

`))
		Expect(string(session.Err.Contents())).To(BeEmpty())
	})

	It("vaultinfo can handle improper yaml", func() {
		session := runSpruce("vaultinfo", "../../assets/vaultinfo/improper.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Out.Contents())).To(BeEmpty())
		Expect(string(session.Err.Contents())).To(Equal(`../../assets/vaultinfo/improper.yml: unmarshal []byte to yaml failed: yaml: line 1: did not find expected node content

`))
	})

	It("Adding (dynamic) prune support for list entries (edge case scenario)", func() {
		session := runSpruce("merge", "../../assets/prune/prune-in-lists/fileA.yml", "../../assets/prune/prune-in-lists/fileB.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`meta:
  list:
  - one
  - three

`))
	})

	It("vaultinfo handles gopatch files", func() {
		session := runSpruce("vaultinfo", "--go-patch", "../../assets/vaultinfo/merge1.yml", "../../assets/vaultinfo/go-patch.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal(`secrets:
- key: secret/beep:boop
  references:
  - bar
- key: secret/blork:blork
  references:
  - new_key
- key: secret/foo:bar
  references:
  - foo

`))
		Expect(string(session.Err.Contents())).To(BeEmpty())
	})

	It("Adding (static) prune support for list entries (edge case scenario)", func() {
		session := runSpruce("merge", "--prune", "meta.list.1", "../../assets/prune/prune-in-lists/fileA.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`meta:
  list:
  - one
  - three

`))
	})

	It("Issue - prune and inject cause side-effect", func() {
		session := runSpruce("merge", "--prune", "meta", "../../assets/prune/prune-issue-with-inject/fileA.yml", "../../assets/prune/prune-issue-with-inject/fileB.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`jobs:
- instances: 2
  name: main-job
  templates:
  - name: one
  - name: two
  update:
    canaries: 1
    max_in_flight: 3
- instances: 1
  name: another-job
  templates:
  - name: one
  - name: two
  update:
    canaries: 2

`))
	})

	It("Issue - prune and new-list-entry cause side-effect", func() {
		session := runSpruce("merge", "--prune", "meta", "../../assets/prune/prune-issue-in-lists-with-new-entry/fileA.yml", "../../assets/prune/prune-issue-in-lists-with-new-entry/fileB.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`list:
- name: A
  release: A
  version: A
- name: B
  release: B
  version: B
- name: C
  release: C
  version: C
- name: D
  release: D

`))
	})

	It("Issue #158 prune doesn't work when goes at the end (regression?) - variant A", func() {
		session := runSpruce("merge", "../../assets/prune/issue-158/test.yml", "../../assets/prune/issue-158/prune.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`test1: t2

`))
	})

	It("Issue #158 prune doesn't work when goes at the end (regression?) - variant B", func() {
		session := runSpruce("merge", "../../assets/prune/issue-158/prune.yml", "../../assets/prune/issue-158/test.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`test1: t2

`))
	})

	It("Text needed (prune/issue-250)", func() {
		session := runSpruce("merge", "../../assets/prune/issue-250/fileA.yml", "../../assets/prune/issue-250/fileB.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`list:
- name: zero
  params:
    fail-fast: false
    preload: true
- name: one
  params:
    fail-fast: false
    preload: true
- name: two
  params:
    preload: false

`))
	})

	It("The delete operator deletes an entry in a simple list", func() {
		session := runSpruce("merge", "../../assets/delete/simple-string-fileA.yml", "../../assets/delete/simple-string-fileB.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`meta:
  list:
  - one
  - two
  - five

`))
	})

	It("The delete operator deletes an entry with whitespaces or special characters in a simple list", func() {
		session := runSpruce("merge", "../../assets/delete/text-fileA.yml", "../../assets/delete/text-fileB.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`meta:
  list:
  - Leonel Messi
  - Oliver Kahn
stuff:
  default_groups:
  - openid
  - cloud_controller.read
  - uaa.user
  - approvals.me
  - profile
  - roles
  - user_attributes
  - uaa.offline_token
  environment_scripts:
  - scripts/configure-HA-hosts.sh
  - scripts/forward_logfiles.sh

`))
	})

	It("Issue #156 Can use concat with static ips", func() {
		session := runSpruce("merge", "../../assets/static_ips/issue-156/concat.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`jobs:
- instances: 1
  name: pepe
  networks:
  - name: cf1
    static_ips:
    - 10.4.5.4
meta:
  network_prefix: "10.4"
networks:
- name: cf1
  subnets:
  - range: 10.4.36.0/24
    static:
    - 10.4.5.4 - 10.4.5.100

`))
	})

	It("Issue #194 Globs with missing sub-items track data flow deps properly", func() {
		session := runSpruce("merge", "../../assets/static_ips/vips-plus-grab.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`jobs:
- instances: 1
  name: bosh
  networks:
  - name: stuff
    static_ips:
    - 1.2.3.4
meta:
  ips:
  - 1.2.3.4
networks:
- name: stuff
  subnets:
  - static:
    - 1.2.3.4
- name: stuff2
  type: vip

`))
	})

	Context("Issue #201 - using `azs` instead of `az` in subnets", func() {
		It("jobs in only one zone can see the IPs of all subnets that mentioned that zone", func() {
			session := runSpruce("merge", "../../assets/static_ips/multi-azs-one-zone-job.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`jobs:
- azs:
  - z1
  instances: 2
  name: static_z1
  networks:
  - name: net1
    static_ips:
    - 10.0.0.1
    - 10.1.1.1
networks:
- name: net1
  subnets:
  - azs:
    - z1
    - z2
    - z3
    static:
    - 10.0.0.1 - 10.0.0.15
  - azs:
    - z1
    static:
    - 10.1.1.1
  - azs:
    - z2
    static:
    - 10.2.2.2

`))
		})

		It("jobs in multiple zones can see the IPs of all subnets mentioning those zones", func() {
			session := runSpruce("merge", "../../assets/static_ips/multi-azs-multi-zone-job.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`jobs:
- azs:
  - z1
  - z2
  - z3
  instances: 2
  name: static_z1
  networks:
  - name: net1
    static_ips:
    - 10.1.1.1
    - 10.2.2.2
networks:
- name: net1
  subnets:
  - azs:
    - z1
    - z2
    - z3
    static:
    - 10.0.0.1 - 10.0.0.15
  - azs:
    - z1
    static:
    - 10.1.1.1
  - azs:
    - z2
    static:
    - 10.2.2.2

`))
		})

		It("a z2-only job cannot see z1-only IPs", func() {
			session := runSpruce("merge", "../../assets/static_ips/multi-azs-z2-underprovision.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal(`1 error(s) detected:
 - $.jobs.static_z1.networks.net1.static_ips: request for static_ip(15) in a pool of only 15 (zero-indexed) static addresses


`))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})

		It("jobs with multiple zones see one copy of available IPs, rather than one copy per zone", func() {
			session := runSpruce("merge", "../../assets/static_ips/multi-azs-multi-underprovision.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal(`1 error(s) detected:
 - $.jobs.static_z1.networks.net1.static_ips: request for static_ip(16) in a pool of only 16 (zero-indexed) static addresses


`))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})

		It("edge case - same index used for different IPs with multi-az subnets", func() {
			session := runSpruce("merge", "../../assets/static_ips/multi-azs-same-index-different-ip.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`jobs:
- azs:
  - z1
  instances: 1
  name: static_z1
  networks:
  - name: net1
    static_ips:
    - 10.1.1.1
- azs:
  - z2
  instances: 1
  name: static_z2
  networks:
  - name: net1
    static_ips:
    - 10.2.2.2
networks:
- name: net1
  subnets:
  - azs:
    - z1
    - z2
    - z3
    static:
    - 10.0.0.1 - 10.0.0.15
  - azs:
    - z1
    static:
    - 10.1.1.1
  - azs:
    - z2
    static:
    - 10.2.2.2

`))
		})

		It("edge case - dont give out same IP when specified in jobs with different zones", func() {
			session := runSpruce("merge", "../../assets/static_ips/multi-azs-same-ip-different-zones.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal(`1 error(s) detected:
 - $.jobs.static_z2.networks.net1.static_ips: tried to use IP '10.0.0.15', but that address is already allocated to static_z1/0


`))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})

		It("edge case - don't give out same IP when using different offsets", func() {
			session := runSpruce("merge", "../../assets/static_ips/multi-azs-same-ip-different-index.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal(`1 error(s) detected:
 - $.jobs.static_z2.networks.net1.static_ips: tried to use IP '10.2.2.2', but that address is already allocated to static_z1/0


`))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})
	})

	Context("Empty operator works", func() {
		expectedOutput := `meta:
  first: {}
  second: []
  third: ""

`
		Context("when merging over maps", func() {
			It("with references as the type", func() {
				session := runSpruce("merge", "../../assets/empty/base.yml", "../../assets/empty/references.yml")
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(string(session.Out.Contents())).To(Equal(expectedOutput))
			})

			It("with literals as the type", func() {
				session := runSpruce("merge", "../../assets/empty/base.yml", "../../assets/empty/literals.yml")
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(string(session.Out.Contents())).To(Equal(expectedOutput))
			})
		})

		Context("when merging over nothing", func() {
			It("with references as the type", func() {
				session := runSpruce("merge", "../../assets/empty/references.yml")
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(string(session.Out.Contents())).To(Equal(expectedOutput))
			})

			It("with literals as the type", func() {
				session := runSpruce("merge", "../../assets/empty/literals.yml")
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(string(session.Out.Contents())).To(Equal(expectedOutput))
			})
		})
	})

	Context("Join operator works", func() {
		It("when dependencies could cause improper evaluation order", func() {
			session := runSpruce("merge", "../../assets/join/issue-155/deps.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`b:
- hello
- world
greeting: hello
output:
- hello world
- hello bye
z:
- hello
- bye

`))
		})
	})

	Context("Calc operator works", func() {
		It("Calc comes with built-in functions", func() {
			session := runSpruce("merge", "--prune", "meta", "../../assets/calc/functions.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`properties:
  homework:
    ceil: 9
    floor: 3
    max: 8.333
    min: 3.666
    mod: 1.001
    pow: 2374.9685
    sqrt: 2.8866937

`))
		})

		It("Calc works with dependencies", func() {
			session := runSpruce("merge", "--prune", "meta", "../../assets/calc/dependencies.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`jobs:
- instances: 4
  name: big_ones
- instances: 1
  name: small_ones
- instances: 2
  name: extra_ones

`))
		})

		It("Calc expects only one argument which is a quoted mathematical expression", func() {
			session := runSpruce("merge", "--prune", "meta", "../../assets/calc/wrong-syntax.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal(`2 error(s) detected:
 - $.jobs.one.instances: calc operator only expects one argument containing the expression
 - $.jobs.two.instances: calc operator argument is suppose to be a quoted mathematical expression (type Literal)


`))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})

		It("Calc operator does not support named variables", func() {
			session := runSpruce("merge", "--prune", "meta", "../../assets/calc/no-named-variables.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal(`1 error(s) detected:
 - $.jobs.one.instances: calc operator does not support named variables in expression: pi, r


`))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})

		It("Calc operator checks input for built-in functions", func() {
			session := runSpruce("merge", "--prune", "meta", "../../assets/calc/bad-functions.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal(`7 error(s) detected:
 - $.properties.homework.ceil: ceil function expects one argument of type float64
 - $.properties.homework.floor: floor function expects one argument of type float64
 - $.properties.homework.max: max function expects two arguments of type float64
 - $.properties.homework.min: min function expects two arguments of type float64
 - $.properties.homework.mod: mod function expects two arguments of type float64
 - $.properties.homework.pow: pow function expects two arguments of type float64
 - $.properties.homework.sqrt: sqrt function expects one argument of type float64


`))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})

		It("Calc operator checks referenced types", func() {
			session := runSpruce("merge", "--prune", "meta", "../../assets/calc/wrong-type.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal(`4 error(s) detected:
 - $.properties.homework.list: path meta.list is of type slice, which cannot be used in calculations
 - $.properties.homework.map: path meta.map is of type map, which cannot be used in calculations
 - $.properties.homework.nil: path meta.nil references a nil value, which cannot be used in calculations
 - $.properties.homework.string: path meta.string is of type string, which cannot be used in calculations


`))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})

		It("Calc returns int64s if possible", func() {
			session := runSpruce("merge", "--prune", "meta", "../../assets/calc/large-ints.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`float: 7.776e+06
int: 7776000

`))
		})
	})

	It("YAML output is ordered the same way each time (#184)", func() {
		expectedOutput := `properties:
  cc:
    quota_definitions:
      q2GB:
        non_basic_services_allowed: true
      q4GB:
        non_basic_services_allowed: true
      q256MB:
        non_basic_services_allowed: true

`
		for i := 0; i < 30; i++ {
			session := runSpruce("merge", "../../assets/output-order/sample.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(expectedOutput))
		}
	})

	Context("Sort test cases", func() {
		It("sort operator functionality", func() {
			session := runSpruce("merge", "../../assets/sort/base.yml", "../../assets/sort/op.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`float_list:
- 1.42
- 2.42
- 3.42
- 4.42
- 5.42
- 6.42
- 7.42
- 8.42
- 9.42
foobar_list:
- foobar: item-6
- foobar: item-7
- foobar: item-8
- foobar: item-9
- foobar: item-g
- foobar: item-h
- foobar: item-i
- foobar: item-j
- foobar: item-k
- foobar: item-l
- foobar: item-m
int_list:
- 1
- 2
- 3
- 4
- 5
- 6
- 7
- 8
- 9
key_list:
- key: item-1
- key: item-2
- key: item-3
- key: item-4
- key: item-a
- key: item-b
- key: item-c
- key: item-d
- key: item-e
- key: item-f
- key: item-g
- key: item-h
- key: item-i
name_list:
- name: item-1
- name: item-2
- name: item-3
- name: item-4
- name: item-5
- name: item-6
- name: item-7
- name: item-8
- name: item-9
- name: item-a
- name: item-b
- name: item-c
- name: item-d
- name: item-e
- name: item-f
- name: item-g
- name: item-h
- name: item-i
- name: item-j
- name: item-k
- name: item-l
- name: item-m
- name: item-n
- name: item-o
- name: item-p
- name: item-q
- name: item-r
- name: item-s
- name: item-t
- name: item-u
- name: item-v
- name: item-w
- name: item-x
- name: item-y
- name: item-z

`))
		})
	})

	Context("Given a Spruce merge using the (( load <location> )) operator", func() {
		Context("When the location is a local location", func() {
			It("The local data (via literal) should be loaded and inserted", func() {
				session := runSpruceWithEnv([]string{"SPRUCE_FILE_BASE_PATH=../../"}, "merge", "../../assets/load/base-local.yml")
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(string(session.Out.Contents())).To(Equal(`yet:
  another:
    yaml:
      structure:
        load:
          complex-list:
          - name: one
          - name: two
          map:
            key: value
          simple-list:
          - one
          - two

`))
			})

			It("Absolute paths are not interpreted as remote locations", func() {
				file, fileErr := os.CreateTemp("../../assets/load", "base-local-abs.yml")
				Expect(fileErr).NotTo(HaveOccurred())
				defer os.Remove(file.Name())

				path, pathErr := filepath.Abs("../../assets/load/users.yml")
				Expect(pathErr).NotTo(HaveOccurred())

				content := "params:\n  users: (( load \"" + path + "\" ))"
				_, err := file.Write([]byte(content))
				Expect(err).NotTo(HaveOccurred())
				Expect(file.Close()).NotTo(HaveOccurred())

				session := runSpruceWithEnv([]string{"SPRUCE_FILE_BASE_PATH=../../"}, "merge", "--prune", "meta", file.Name())
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(string(session.Out.Contents())).To(Equal(`params:
  users:
  - color: green
    name: bob
  - color: red
    name: fred

`))
			})

			It("The local data (via reference) should be loaded and inserted", func() {
				session := runSpruce("merge", "--prune", "meta", "../../assets/load/base-local-ref.yml")
				Eventually(session, "10s").Should(gexec.Exit(0))
				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(string(session.Out.Contents())).To(Equal(`params:
  users:
  - color: green
    name: bob
  - color: red
    name: fred

`))
			})

			It("That an error is returned if no file can be found", func() {
				session := runSpruce("merge", "../../assets/load/base-local.yml")
				Eventually(session, "10s").Should(gexec.Exit(2))
				Expect(string(session.Err.Contents())).To(Equal(`1 error(s) detected:
 - $.yet.another.yaml.structure.load: unable to get any content using location assets/load/other.yml: it is not a file or usable URI


`))
				Expect(string(session.Out.Contents())).To(BeEmpty())
			})
		})

		Context("When the location is a remote location", func() {
			var srv *http.Server

			BeforeEach(func() {
				mux := http.NewServeMux()
				mux.Handle("/assets/",
					http.StripPrefix("/assets/",
						http.FileServer(http.Dir("../../assets/"))))
				srv = &http.Server{Addr: ":31337", Handler: mux}
				go func() {
					srv.ListenAndServe() //nolint:errcheck
				}()
				time.Sleep(500 * time.Millisecond)
			})

			AfterEach(func() {
				if srv != nil {
					srv.Shutdown(context.Background()) //nolint:errcheck
				}
			})

			It("The remote data should be loaded and inserted", func() {
				session := runSpruce("merge", "../../assets/load/base-remote.yml")
				Eventually(session, "15s").Should(gexec.Exit(0))
				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(string(session.Out.Contents())).To(Equal(`yet:
  another:
    yaml:
      structure:
        load:
        - one
        - two

`))
			})
		})
	})

	Context("Cherry picking test cases", func() {
		It("Cherry pick just one root level path", func() {
			session := runSpruce("merge", "--cherry-pick", "properties", "../../assets/cherry-pick/fileA.yml", "../../assets/cherry-pick/fileB.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`properties:
  hahn:
    flags: open
    id: b503e54a-c872-4643-a09c-5480c5940d0c
  vb:
    flags: auth,block,read-only
    id: 74a03820-3f81-45ca-afd5-d7d57b947ff1

`))
		})

		It("Cherry pick a path that is a list entry", func() {
			session := runSpruce("merge", "--cherry-pick", "releases.vb", "../../assets/cherry-pick/fileA.yml", "../../assets/cherry-pick/fileB.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`releases:
- name: vb

`))
		})

		It("Cherry pick a path that is deep down the structure", func() {
			session := runSpruce("merge", "--cherry-pick", "meta.some.deep.structure.maplist", "../../assets/cherry-pick/fileA.yml", "../../assets/cherry-pick/fileB.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`meta:
  some:
    deep:
      structure:
        maplist:
          keyA: valueA
          keyB: valueB

`))
		})

		It("Cherry pick a series of different paths at the same time", func() {
			session := runSpruce("merge", "--cherry-pick", "properties", "--cherry-pick", "releases.vb", "--cherry-pick", "meta.some.deep.structure.maplist", "../../assets/cherry-pick/fileA.yml", "../../assets/cherry-pick/fileB.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`meta:
  some:
    deep:
      structure:
        maplist:
          keyA: valueA
          keyB: valueB
properties:
  hahn:
    flags: open
    id: b503e54a-c872-4643-a09c-5480c5940d0c
  vb:
    flags: auth,block,read-only
    id: 74a03820-3f81-45ca-afd5-d7d57b947ff1
releases:
- name: vb

`))
		})

		It("Cherry pick a path and prune something at the same time in a map", func() {
			session := runSpruce("merge", "--cherry-pick", "properties", "--prune", "properties.vb.flags", "../../assets/cherry-pick/fileA.yml", "../../assets/cherry-pick/fileB.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`properties:
  hahn:
    flags: open
    id: b503e54a-c872-4643-a09c-5480c5940d0c
  vb:
    id: 74a03820-3f81-45ca-afd5-d7d57b947ff1

`))
		})

		It("Cherry picking should fail if you cherry-pick a prune path", func() {
			session := runSpruce("merge", "--cherry-pick", "properties", "--prune", "properties", "../../assets/cherry-pick/fileA.yml", "../../assets/cherry-pick/fileB.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal("1 error(s) detected:\n" +
				" - `$.properties` could not be found in the datastructure\n\n\n"))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})

		It("Cherry picking should fail if picking a sub-level path while prune wipes the parent", func() {
			session := runSpruce("merge", "--cherry-pick", "releases.vb", "--prune", "releases", "../../assets/cherry-pick/fileA.yml", "../../assets/cherry-pick/fileB.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal("1 error(s) detected:\n" +
				" - `$.releases` could not be found in the datastructure\n\n\n"))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})

		It("Cherry pick a list entry path of a list that uses 'key' as its identifier", func() {
			session := runSpruce("merge", "--cherry-pick", "list.two", "../../assets/cherry-pick/key-based-list.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`list:
- desc: The second one
  key: two
  version: v2

`))
		})

		It("Cherry pick a list entry path of a list that uses 'id' as its identifier", func() {
			session := runSpruce("merge", "--cherry-pick", "list.two", "../../assets/cherry-pick/id-based-list.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`list:
- desc: The second one
  id: two
  version: v2

`))
		})

		It("Cherry pick one list entry path that references the index", func() {
			session := runSpruce("merge", "--cherry-pick", "list.1", "../../assets/cherry-pick/name-based-list.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`list:
- desc: The second one
  name: two
  version: v2

`))
		})

		It("Cherry pick two list entry paths that reference indexes", func() {
			session := runSpruce("merge", "--cherry-pick", "list.1", "--cherry-pick", "list.4", "../../assets/cherry-pick/name-based-list.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`list:
- desc: The fifth one
  name: five
  version: v5
- desc: The second one
  name: two
  version: v2

`))
		})

		It("Cherry pick one list entry path that references an invalid index", func() {
			session := runSpruce("merge", "--cherry-pick", "list.10", "../../assets/cherry-pick/name-based-list.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal("1 error(s) detected:\n" +
				" - `$.list.10` could not be found in the datastructure\n\n\n"))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})

		It("Cherry pick should only pick the exact name based on the path", func() {
			session := runSpruce("merge", "--cherry-pick", "map", "--prune", "subkey", "../../assets/cherry-pick/test-exact-names.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`map:
  other: value
  subkey: this is the real subkey

`))
		})

		It("Cherry pick should only evaluate the dynamic operators that are relevant", func() {
			session := runSpruce("merge", "--cherry-pick", "params", "../../assets/cherry-pick/partial-eval.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`params:
  mode: default
  name: sandbox-thing
  type: thing

`))
		})
	})

	It("FallbackAppend should cause the default behavior after a key merge to go to append", func() {
		session := runSpruce("merge", "--fallback-append", "../../assets/fallback-append/test1.yml", "../../assets/fallback-append/test2.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`array:
- thing: 1
  value: foo
- thing: 2
  value: bar
- thing: 1
  value: baz

`))
	})

	It("Without FallbackAppend, the default merge behavior after a key merge should still be inline", func() {
		session := runSpruce("merge", "../../assets/fallback-append/test1.yml", "../../assets/fallback-append/test2.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`array:
- thing: 1
  value: baz
- thing: 2
  value: bar

`))
	})

	Context("Defer", func() {
		It("should err if there are no arguments", func() {
			session := runSpruce("merge", "../../assets/defer/nothing.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal(`1 error(s) detected:
 - $.foo: defer has no arguments - what are you deferring?


`))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})

		It("on a non-quoted string", func() {
			session := runSpruce("merge", "../../assets/defer/simple-ref.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`foo: (( thing ))

`))
		})

		It("on a quoted string", func() {
			session := runSpruce("merge", "../../assets/defer/simple-string.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`foo: (( "thing" ))

`))
		})

		It("on a non-quoted string called nil", func() {
			session := runSpruce("merge", "../../assets/defer/simple-nil.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`foo: (( nil ))

`))
		})

		It("on an integer", func() {
			session := runSpruce("merge", "../../assets/defer/simple-int.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`foo: (( 123 ))

`))
		})

		It("on a float", func() {
			session := runSpruce("merge", "../../assets/defer/simple-float.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`foo: (( 1.23 ))

`))
		})

		It("on an environment variable", func() {
			session := runSpruce("merge", "../../assets/defer/simple-envvar.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`foo: (( $TESTVAR ))

`))
		})

		It("on an unquoted string that could reference another key", func() {
			session := runSpruce("merge", "../../assets/defer/reference.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`foo: (( thing ))
thing: (( thing ))

`))
		})

		It("on a value with a logical-or", func() {
			session := runSpruce("merge", "../../assets/defer/or.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`foo: (( grab this || "that" ))

`))
		})

		It("with another operator in the defer", func() {
			session := runSpruce("merge", "../../assets/defer/grab.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`foo: (( grab thing ))
grab: beep
thing: boop

`))
		})
	})

	Context("non-specific node tags specific test cases", func() {
		It("non-specific node tags test case - style 1", func() {
			session := runSpruce("merge", "../../assets/non-specific-node-tags-issue/fileA-1.yml", "../../assets/non-specific-node-tags-issue/fileB.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`some:
  yaml:
    structure:
      certificate: |
        -----BEGIN CERTIFICATE-----
        QSBzcHJ1Y2UgaXMgYSB0cmVlIG9mIHRoZSBnZW51cyBQaWNlYSAvcGHJqsuIc2nL
        kMmZLyxbMV0gYSBnZW51cyBvZiBhYm91dCAzNSBzcGVjaWVzIG9mIGNvbmlmZXJv
        dXMgZXZlcmdyZWVuIHRyZWVzIGluIHRoZSBGYW1pbHkgUGluYWNlYWUsIGZvdW5k
        IGluIHRoZSBub3J0aGVybiB0ZW1wZXJhdGUgYW5kIGJvcmVhbCAodGFpZ2EpIHJl
        Z2lvbnMgb2YgdGhlIGVhcnRoLiBTcHJ1Y2VzIGFyZSBsYXJnZSB0cmVlcywgZnJv
        bSBhYm91dCAyMOKAkzYwIG1ldHJlcyAoYWJvdXQgNjDigJMyMDAgZmVldCkgdGFs
        bCB3aGVuIG1hdHVyZSwgYW5kIGNhbiBiZSBkaXN0aW5ndWlzaGVkIGJ5IHRoZWly
        IHdob3JsZWQgYnJhbmNoZXMgYW5kIGNvbmljYWwgZm9ybS4gVGhlIG5lZWRsZXMs
        IG9yIGxlYXZlcywgb2Ygc3BydWNlIHRyZWVzIGFyZSBhdHRhY2hlZCBzaW5nbHkg
        dG8gdGhlIGJyYW5jaGVzIGluIGEgc3BpcmFsIGZhc2hpb24sIGVhY2ggbmVlZGxl
        IG9uIGEgc21hbGwgcGVnLWxpa2Ugc3RydWN0dXJlLiBUaGUgbmVlZGxlcyBhcmUg
        c2hlZCB3aGVuIDTigJMxMCB5ZWFycyBvbGQsIGxlYXZpbmcgdGhlIGJyYW5jaGVz
        IHJvdWdoIHdpdGggdGhlIHJldGFpbmVkIHBlZ3MgKGFuIGVhc3kgbWVhbnMgb2Yg
        ZGlzdGluZ3Vpc2hpbmcgdGhlbSBmcm9tIG90aGVyIHNpbWlsYXIgZ2VuZXJhLCB3
        aGVyZSB0aGUgYnJhbmNoZXMgYXJlIGZhaXJseSBzbW9vdGgpLgoKU3BydWNlcyBh
        cmUgdXNlZCBhcyBmb29kIHBsYW50cyBieSB0aGUgbGFydmFlIG9mIHNvbWUgTGVw
        aWRvcHRlcmEgKG1vdGggYW5kIGJ1dHRlcmZseSkgc3BlY2llczsgc2VlIGxpc3Qg
        b2YgTGVwaWRvcHRlcmEgdGhhdCBmZWVkIG9uIHNwcnVjZXMuIFRoZXkgYXJlIGFs
        c28gdXNlZCBieSB0aGUgbGFydmFlIG9mIGdhbGwgYWRlbGdpZHMgKEFkZWxnZXMg
        c3BlY2llcykuCgpJbiB0aGUgbW91bnRhaW5zIG9mIHdlc3Rlcm4gU3dlZGVuIHNj
        aWVudGlzdHMgaGF2ZSBmb3VuZCBhIE5vcndheSBzcHJ1Y2UgdHJlZSwgbmlja25h
        bWVkIE9sZCBUamlra28sIHdoaWNoIGJ5IHJlcHJvZHVjaW5nIHRocm91Z2ggbGF5
        ZXJpbmcgaGFzIHJlYWNoZWQgYW4gYWdlIG9mIDksNTUwIHllYXJzIGFuZCBpcyBj
        bGFpbWVkIHRvIGJlIHRoZSB3b3JsZCdzIG9sZGVzdCBrbm93biBsaXZpbmcgdHJl
        ZS4K
        -----END CERTIFICATE-----
      someotherkey: value

`))
		})

		It("non-specific node tags test case - style 2", func() {
			session := runSpruce("merge", "../../assets/non-specific-node-tags-issue/fileA-2.yml", "../../assets/non-specific-node-tags-issue/fileB.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`some:
  yaml:
    structure:
      certificate: '-----BEGIN CERTIFICATE----- QSBzcHJ1Y2UgaXMgYSB0cmVlIG9mIHRoZSBnZW51cyBQaWNlYSAvcGHJqsuIc2nL
        kMmZLyxbMV0gYSBnZW51cyBvZiBhYm91dCAzNSBzcGVjaWVzIG9mIGNvbmlmZXJv dXMgZXZlcmdyZWVuIHRyZWVzIGluIHRoZSBGYW1pbHkgUGluYWNlYWUsIGZvdW5k
        IGluIHRoZSBub3J0aGVybiB0ZW1wZXJhdGUgYW5kIGJvcmVhbCAodGFpZ2EpIHJl Z2lvbnMgb2YgdGhlIGVhcnRoLiBTcHJ1Y2VzIGFyZSBsYXJnZSB0cmVlcywgZnJv
        bSBhYm91dCAyMOKAkzYwIG1ldHJlcyAoYWJvdXQgNjDigJMyMDAgZmVldCkgdGFs bCB3aGVuIG1hdHVyZSwgYW5kIGNhbiBiZSBkaXN0aW5ndWlzaGVkIGJ5IHRoZWly
        IHdob3JsZWQgYnJhbmNoZXMgYW5kIGNvbmljYWwgZm9ybS4gVGhlIG5lZWRsZXMs IG9yIGxlYXZlcywgb2Ygc3BydWNlIHRyZWVzIGFyZSBhdHRhY2hlZCBzaW5nbHkg
        dG8gdGhlIGJyYW5jaGVzIGluIGEgc3BpcmFsIGZhc2hpb24sIGVhY2ggbmVlZGxl IG9uIGEgc21hbGwgcGVnLWxpa2Ugc3RydWN0dXJlLiBUaGUgbmVlZGxlcyBhcmUg
        c2hlZCB3aGVuIDTigJMxMCB5ZWFycyBvbGQsIGxlYXZpbmcgdGhlIGJyYW5jaGVz IHJvdWdoIHdpdGggdGhlIHJldGFpbmVkIHBlZ3MgKGFuIGVhc3kgbWVhbnMgb2Yg
        ZGlzdGluZ3Vpc2hpbmcgdGhlbSBmcm9tIG90aGVyIHNpbWlsYXIgZ2VuZXJhLCB3 aGVyZSB0aGUgYnJhbmNoZXMgYXJlIGZhaXJseSBzbW9vdGgpLgoKU3BydWNlcyBh
        cmUgdXNlZCBhcyBmb29kIHBsYW50cyBieSB0aGUgbGFydmFlIG9mIHNvbWUgTGVw aWRvcHRlcmEgKG1vdGggYW5kIGJ1dHRlcmZseSkgc3BlY2llczsgc2VlIGxpc3Qg
        b2YgTGVwaWRvcHRlcmEgdGhhdCBmZWVkIG9uIHNwcnVjZXMuIFRoZXkgYXJlIGFs c28gdXNlZCBieSB0aGUgbGFydmFlIG9mIGdhbGwgYWRlbGdpZHMgKEFkZWxnZXMg
        c3BlY2llcykuCgpJbiB0aGUgbW91bnRhaW5zIG9mIHdlc3Rlcm4gU3dlZGVuIHNj aWVudGlzdHMgaGF2ZSBmb3VuZCBhIE5vcndheSBzcHJ1Y2UgdHJlZSwgbmlja25h
        bWVkIE9sZCBUamlra28sIHdoaWNoIGJ5IHJlcHJvZHVjaW5nIHRocm91Z2ggbGF5 ZXJpbmcgaGFzIHJlYWNoZWQgYW4gYWdlIG9mIDksNTUwIHllYXJzIGFuZCBpcyBj
        bGFpbWVkIHRvIGJlIHRoZSB3b3JsZCdzIG9sZGVzdCBrbm93biBsaXZpbmcgdHJl ZS4K -----END
        CERTIFICATE-----'
      someotherkey: value

`))
		})

		It("Issue #198 - avoid nil panics when merging arrays with nil elements", func() {
			session := runSpruce("merge", "../../assets/issue-198/nil-array-elements.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`empty_nil:
- null
- more stuff
explicit_nil:
- null
- stuff
latter_elements_nil:
- stuff
- null
nested_nil:
- stuff:
  - null
  - nested nil above
  thing: has stuff

`))
		})

		It("Issue #172 - don't panic if target key has map value", func() {
			session := runSpruce("merge", "../../assets/issue-172/implicitmergemap.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(Equal(`warning: $.array-of-maps.0: new object's key 'name' cannot have a value which is a hash or sequence - cannot merge by key
warning: Falling back to inline merge strategy
`))
			Expect(string(session.Out.Contents())).To(Equal(`array-of-maps:
- name:
    subkey1: true
    subkey2: false

`))
		})

		It("Issue #172 - don't panic if target key has sequence value", func() {
			session := runSpruce("merge", "../../assets/issue-172/implicitmergeseq.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(Equal(`warning: $.array-of-maps.0: new object's key 'name' cannot have a value which is a hash or sequence - cannot merge by key
warning: Falling back to inline merge strategy
`))
			Expect(string(session.Out.Contents())).To(Equal(`array-of-maps:
- name:
  - subkey1
  - subkey2

`))
		})

		It("Issue #172 - error instead of panic if merge was specifically requested but target key has map value", func() {
			session := runSpruce("merge", "../../assets/issue-172/explicitmerge1.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal(`1 error(s) detected:
 - $.array-of-maps.0: new object's key 'name' cannot have a value which is a hash or sequence - cannot merge by key


`))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})

		It("Issue #172 - error instead of panic if merge on key was specifically requested but target key has map value", func() {
			session := runSpruce("merge", "../../assets/issue-172/explicitmergeonkey1.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal(`1 error(s) detected:
 - $.array-of-maps.0: new object's key 'mergekey' cannot have a value which is a hash or sequence - cannot merge by key


`))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})
	})

	Context("Issue #215 - Handle really big ints as operator arguments", func() {
		It("We didn't break normal small ints", func() {
			session := runSpruce("merge", "../../assets/issue-215/smallint.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal("foo: -> 3 <-\n\n"))
		})

		It("We can handle ints bigger than 2^63 - 1", func() {
			session := runSpruce("merge", "../../assets/issue-215/hugeint.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal("foo: -> 6.239871649276491e+24 <-\n\n"))
		})
	})

	It("Issue #153 - Cartesian Product should produce a []interface{}", func() {
		session := runSpruce("merge", "../../assets/cartesian-product/can-be-joined.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`ips:
- 1.2.3.4
- 2.2.3.4
ips_with_port:
- 1.2.3.4:80
- 2.2.3.4:80
join_ips_with_port: 1.2.3.4:80,2.2.3.4:80

`))
	})

	It("Issue #169 - Cartesian Product should produce a []interface{}", func() {
		session := runSpruce("merge", "../../assets/cartesian-product/can-be-grabbed.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`groups:
- jobs:
  - master-isolation-tests
  - master-integration-tests
  - master-dependencies-test
  - master-docker-build
  name: master
meta:
  fast-tests:
  - isolation
  master-fast-tests:
  - master-isolation-tests
  master-slow-tests:
  - master-integration-tests
  slow-tests:
  - integration

`))
	})

	Context("Issue #267 - specifying an explicit merge operator must behave in the same way as relying on the default implicit merge operation", func() {
		It("Option 1 - standard use-case: no explicit merge, named-entry list identifier key is the default called 'name'", func() {
			session := runSpruce("merge", "../../assets/issue-267/option1-fileA.yml", "../../assets/issue-267/option1-fileB.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`serverFiles:
  prometheus.yml:
    scrape_configs:
    - name: one
    - name: two

`))
		})

		It("Option 2 - academic version of the option 1: same set-up, but with explicit usage of the merge operator", func() {
			session := runSpruce("merge", "../../assets/issue-267/option2-fileA.yml", "../../assets/issue-267/option2-fileB.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`serverFiles:
  prometheus.yml:
    scrape_configs:
    - name: one
    - name: two

`))
		})

		It("Option 3 - even more academic version with explicit merge and default identifier key", func() {
			session := runSpruce("merge", "../../assets/issue-267/option3-fileA.yml", "../../assets/issue-267/option3-fileB.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`serverFiles:
  prometheus.yml:
    scrape_configs:
    - name: one
    - name: two

`))
		})

		It("Option 4 - actual real world use case where the identifier key is 'job_name'", func() {
			session := runSpruce("merge", "../../assets/issue-267/option4-fileA.yml", "../../assets/issue-267/option4-fileB.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`serverFiles:
  prometheus.yml:
    scrape_configs:
    - job_name: one
    - job_name: two

`))
		})
	})

	Context("Support go-patch files", func() {
		It("go-patch can modify yaml files in the merge phase, and insert spruce operators as required", func() {
			session := runSpruce("merge", "--go-patch", "../../assets/go-patch/base.yml", "../../assets/go-patch/patch.yml", "../../assets/go-patch/toMerge.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`array:
- 10
- 5
- 6
items:
- add spruce stuff in the beginning of the array
- name: item7
- name: item8
- name: item9
key: 10
key2:
  nested:
    another_nested:
      super_nested: 10
    super_nested: 10
  other: 3
more_stuff: is here
new_key: 10
spruce_array_grab:
- add spruce stuff in the beginning of the array
- name: item7
- name: item8
- name: item9

`))
		})

		It("go-patch throws errors to the front-end when there are go-patch issues", func() {
			session := runSpruce("merge", "--go-patch", "../../assets/go-patch/base.yml", "../../assets/go-patch/err.yml", "../../assets/go-patch/toMerge.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(Equal(`../../assets/go-patch/err.yml: Expected to find a map key 'key_not_there' for path '/key_not_there' (found map keys: 'array', 'items', 'key', 'key2')

`))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})

		It("yaml-parser throws errors when trying to parse gopatch from array-based files", func() {
			session := runSpruce("merge", "--go-patch", "../../assets/go-patch/base.yml", "../../assets/go-patch/bad.yml")
			Eventually(session, "10s").Should(gexec.Exit(2))
			Expect(string(session.Err.Contents())).To(ContainSubstring("Root of YAML document is not a hash/map. Tried parsing it as go-patch, but got:"))
			Expect(string(session.Out.Contents())).To(BeEmpty())
		})

		It("go-patch handles named arrays with :before syntax (#283)", func() {
			session := runSpruce("merge", "--go-patch", "../../assets/go-patch/base.yml", "../../assets/go-patch/before.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`array:
- 4
- 5
- 6
items:
- name: item7
- name: 7.5
- name: item8
- name: item9
key: 1
key2:
  nested:
    super_nested: 2
  other: 3

`))
		})
	})

	Context("setting DEFAULT_ARRAY_MERGE_KEY", func() {
		It("changes how arrays of maps are merged by default", func() {
			session := runSpruceWithEnv([]string{"DEFAULT_ARRAY_MERGE_KEY=id"}, "merge", "../../assets/default-array-merge-var/first.yml", "../../assets/default-array-merge-var/second.yml")
			Eventually(session, "10s").Should(gexec.Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(string(session.Out.Contents())).To(Equal(`array:
- id: first
  value: 123
- id: second
  value: 987
- id: third
  value: true

`))
		})
	})
})

var _ = Describe("debug flags", func() {
	It("-D flag is accepted without error (exits with usage code)", func() {
		session := runSpruce("-D")
		Eventually(session, "10s").Should(gexec.Exit(1))
	})

	It("--debug flag is accepted without error (exits with usage code)", func() {
		session := runSpruce("--debug")
		Eventually(session, "10s").Should(gexec.Exit(1))
	})

	It("DEBUG=tRuE enables debug mode (binary starts without crash)", func() {
		session := runSpruceWithEnv([]string{"DEBUG=tRuE"})
		Eventually(session, "10s").Should(gexec.Exit(1))
	})

	It("DEBUG=1 enables debug mode (binary starts without crash)", func() {
		session := runSpruceWithEnv([]string{"DEBUG=1"})
		Eventually(session, "10s").Should(gexec.Exit(1))
	})

	It("DEBUG=randomval enables debug mode (binary starts without crash)", func() {
		session := runSpruceWithEnv([]string{"DEBUG=randomval"})
		Eventually(session, "10s").Should(gexec.Exit(1))
	})

	It("DEBUG=fAlSe disables debugging (binary starts without crash)", func() {
		session := runSpruceWithEnv([]string{"DEBUG=fAlSe"})
		Eventually(session, "10s").Should(gexec.Exit(1))
	})

	It("DEBUG=0 disables debugging (binary starts without crash)", func() {
		session := runSpruceWithEnv([]string{"DEBUG=0"})
		Eventually(session, "10s").Should(gexec.Exit(1))
	})

	It("DEBUG=empty disables debugging (binary starts without crash)", func() {
		session := runSpruceWithEnv([]string{"DEBUG="})
		Eventually(session, "10s").Should(gexec.Exit(1))
	})
})

var _ = Describe("spruce fan", func() {
	It("errors when failing to read a file it was given", func() {
		session := runSpruce("fan", "../../assets/fan/nonexistent.yml", "../../assets/fan/multi-doc-1.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).To(ContainSubstring("Error reading file ../../assets/fan/nonexistent.yml: open ../../assets/fan/nonexistent.yml: no such file or directory"))
		Expect(string(session.Out.Contents())).To(BeEmpty())
	})

	It("errors with the correct document index when there's an initial doc-separator", func() {
		session := runSpruce("fan", "../../assets/fan/source.yml", "../../assets/fan/invalid-yaml-with-doc-separator.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).To(ContainSubstring("../../assets/fan/invalid-yaml-with-doc-separator.yml[0]:"))
		Expect(string(session.Out.Contents())).To(BeEmpty())
	})

	It("errors with the correct doc index when there is no initial doc separator", func() {
		session := runSpruce("fan", "../../assets/fan/source.yml", "../../assets/fan/invalid-yaml.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).To(ContainSubstring("../../assets/fan/invalid-yaml.yml[0]:"))
		Expect(string(session.Out.Contents())).To(BeEmpty())
	})

	It("errors if no source file is provided", func() {
		session := runSpruce("fan")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).To(ContainSubstring("You must specify at least a source document to spruce fan. If no files are specified, STDIN is used. Using STDIN for source and target docs only works with -m"))
		Expect(string(session.Out.Contents())).To(BeEmpty())
	})

	It("merges one doc into all the docs of the other files", func() {
		session := runSpruce("fan", "--prune", "meta", "../../assets/fan/source.yml", "../../assets/fan/multi-doc-1.yml", "../../assets/fan/multi-doc-2.yml", "../../assets/fan/multi-doc-3.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`---
doc1: i've-been-grabbed

---
doc2: i've-been-grabbed
other: stuff

---
doc3: i've-been-grabbed

---
no-grab: here

---
doc4: i've-been-grabbed

---
doc5: i've-been-grabbed
other: stuff

---
doc6:
  no-grab: here

---
doc7:
  no-grab: here

`))
	})

	It("merges a multi doc source into all the docs of the other files", func() {
		session := runSpruce("fan", "-m", "--prune", "meta", "../../assets/fan/multi-doc-source.yml", "../../assets/fan/multi-doc-1.yml", "../../assets/fan/multi-doc-2.yml", "../../assets/fan/multi-doc-3.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).To(Equal(`---
sdoc: i've-been-grabbed

---
doc1: i've-been-grabbed

---
doc2: i've-been-grabbed
other: stuff

---
doc3: i've-been-grabbed

---
no-grab: here

---
doc4: i've-been-grabbed

---
doc5: i've-been-grabbed
other: stuff

---
doc6:
  no-grab: here

---
doc7:
  no-grab: here

`))
	})
})

var _ = Describe("Examples from README.md", func() {
	It("Basic Example", func() {
		session := runSpruce("merge", "../../examples/basic/main.yml", "../../examples/basic/merge.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).NotTo(BeEmpty())
	})

	It("Map Replacements", func() {
		session := runSpruce("merge", "../../examples/map-replacement/original.yml", "../../examples/map-replacement/delete.yml", "../../examples/map-replacement/insert.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
		Expect(string(session.Out.Contents())).NotTo(BeEmpty())
	})

	It("Key Removal", func() {
		session := runSpruce("merge", "--prune", "deleteme", "../../examples/key-removal/original.yml", "../../examples/key-removal/things.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())

		session2 := runSpruce("merge", "../../examples/pruning/base.yml", "../../examples/pruning/jobs.yml", "../../examples/pruning/networks.yml")
		Eventually(session2, "10s").Should(gexec.Exit(0))
		Expect(string(session2.Err.Contents())).To(BeEmpty())
	})

	It("Lists of Maps", func() {
		session := runSpruce("merge", "../../examples/list-of-maps/original.yml", "../../examples/list-of-maps/new.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
	})

	It("Static IPs", func() {
		session := runSpruce("merge", "../../examples/static-ips/jobs.yml", "../../examples/static-ips/properties.yml", "../../examples/static-ips/networks.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
	})

	It("Static IPs with availability zones", func() {
		session := runSpruce("merge", "../../examples/availability-zones/jobs.yml", "../../examples/availability-zones/properties.yml", "../../examples/availability-zones/networks.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
	})

	It("Injecting Subtrees", func() {
		session := runSpruce("merge", "--prune", "meta", "../../examples/inject/all-in-one.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())

		session2 := runSpruce("merge", "--prune", "meta", "../../examples/inject/templates.yml", "../../examples/inject/green.yml")
		Eventually(session2, "10s").Should(gexec.Exit(0))
		Expect(string(session2.Err.Contents())).To(BeEmpty())
	})

	It("Pruning", func() {
		session := runSpruce("merge", "../../examples/pruning/base.yml", "../../examples/pruning/jobs.yml", "../../examples/pruning/networks.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
	})

	It("Inserting", func() {
		session := runSpruce("merge", "../../examples/inserting/main.yml", "../../examples/inserting/addon.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
	})

	It("Calc", func() {
		session := runSpruce("merge", "--prune", "meta", "../../examples/calc/meta.yml", "../../examples/calc/jobs.yml")
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Err.Contents())).To(BeEmpty())
	})
})

var _ = Describe("CLI Integration - stdin merging", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "spruce-cli-test-*")
		Expect(err).NotTo(HaveOccurred())

		err = os.WriteFile(filepath.Join(tmpDir, "first.yml"), []byte("---\nfirst: beginning\n"), 0644)
		Expect(err).NotTo(HaveOccurred())

		err = os.WriteFile(filepath.Join(tmpDir, "last.yml"), []byte("---\nlast: ending\n"), 0644)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	It("merges yaml from stdin when no files specified", func() {
		cmd := exec.Command(sprucePath, "merge")
		cmd.Stdin = strings.NewReader("first: stdin\n")
		cmd.Dir = tmpDir
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal("first: stdin\n\n"))
	})

	It("merges yaml from stdin with explicit '-' arg", func() {
		cmd := exec.Command(sprucePath, "merge", "-")
		cmd.Stdin = strings.NewReader("first: stdin\n")
		cmd.Dir = tmpDir
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal("first: stdin\n\n"))
	})

	It("merges yaml from stdin with first.yml and last.yml (stdin overrides first)", func() {
		cmd := exec.Command(sprucePath, "merge", "first.yml", "-", "last.yml")
		cmd.Stdin = strings.NewReader("first: stdin\n")
		cmd.Dir = tmpDir
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal("first: stdin\nlast: ending\n\n"))
	})

	It("merges yaml from stdin with first.yml and last.yml (last overrides stdin)", func() {
		cmd := exec.Command(sprucePath, "merge", "first.yml", "-", "last.yml")
		cmd.Stdin = strings.NewReader("last: stdin\n")
		cmd.Dir = tmpDir
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(Equal("first: beginning\nlast: ending\n\n"))
	})
})

var _ = Describe("Color Output - error conditions exit with code 2 and produce non-empty stderr", func() {
	It("Bad YAML for 'spruce json'", func() {
		session := runSpruceWithStdin(`"3"`, "json")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).NotTo(BeEmpty())
	})

	It("Empty stdin for 'spruce json' succeeds with no output", func() {
		cmd := exec.Command(sprucePath, "json")
		cmd.Stdin = strings.NewReader("")
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, "10s").Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).To(BeEmpty())
	})

	It("Bad file for 'spruce json'", func() {
		session := runSpruce("json", "nonexistent.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).NotTo(BeEmpty())
	})

	It("Bad YAML root for merge", func() {
		session := runSpruce("merge", "../../assets/json/non-map.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).NotTo(BeEmpty())
	})

	It("Bad File for merge", func() {
		session := runSpruce("merge", "nonexistent.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).NotTo(BeEmpty())
	})

	It("Bad YAML parsing for merge", func() {
		session := runSpruce("merge", "../../assets/json/malformed.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).NotTo(BeEmpty())
	})

	It("Recursion Depth error", func() {
		session := runSpruce("merge", "../../assets/dereference/cyclic-data.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).NotTo(BeEmpty())
	})

	It("Bad merge into non-array", func() {
		session := runSpruce("merge", "../../assets/merge/first.yml", "../../assets/merge/non-array-merge.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).NotTo(BeEmpty())
	})

	It("Bad merge using keys", func() {
		session := runSpruce("merge", "../../assets/merge/first.yml", "../../assets/merge/error.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).NotTo(BeEmpty())
	})

	It("Bad merge using keys that don't exist", func() {
		session := runSpruce("merge", "../../assets/merge/first.yml", "../../assets/merge/no-key-merge.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).NotTo(BeEmpty())
	})

	It("Syntax error in concat", func() {
		session := runSpruce("merge", "../../assets/concat/fail.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).NotTo(BeEmpty())
	})

	It("NotFoundError, TypeMisMatchErrors, all operator errors, all tree errors", func() {
		session := runSpruce("merge", "../../assets/errors/colortest.yml")
		Eventually(session, "10s").Should(gexec.Exit(2))
		Expect(string(session.Err.Contents())).NotTo(BeEmpty())
	})
})

// Ensure fmt is used (for unused import prevention)
var _ = fmt.Sprintf
