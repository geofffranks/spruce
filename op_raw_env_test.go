package spruce

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RawEnv Operator", func() {
	op := RawEnvOperator{}
	ev := &Evaluator{}

	BeforeEach(func() {
		os.Setenv("RAW_ENV_TEST", "hello world")
		os.Setenv("RAW_ENV_SCI", "25e25")
		os.Setenv("RAW_ENV_BOOL", "true")
		os.Setenv("RAW_ENV_NUM", "42")
		os.Setenv("RAW_ENV_FLOAT", "3.14159")
		os.Setenv("RAW_ENV_MULTILINE", "line1\n\nline3\nline4")
		os.Setenv("RAW_ENV_EMPTY_STRING", "")
		os.Setenv("RAW_ENV_YAML_LIST", "[1, 2, 3]")
		os.Setenv("RAW_ENV_YAML_MAP", "{key: value}")
	})

	AfterEach(func() {
		os.Unsetenv("RAW_ENV_TEST")
		os.Unsetenv("RAW_ENV_SCI")
		os.Unsetenv("RAW_ENV_BOOL")
		os.Unsetenv("RAW_ENV_NUM")
		os.Unsetenv("RAW_ENV_FLOAT")
		os.Unsetenv("RAW_ENV_MULTILINE")
		os.Unsetenv("RAW_ENV_EMPTY_STRING")
		os.Unsetenv("RAW_ENV_YAML_LIST")
		os.Unsetenv("RAW_ENV_YAML_MAP")
	})

	It("can retrieve an environment variable as raw string", func() {
		r, err := op.Run(ev, []*Expr{
			env("RAW_ENV_TEST"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("hello world"))
	})

	It("preserves scientific notation as string without converting to float", func() {
		r, err := op.Run(ev, []*Expr{
			env("RAW_ENV_SCI"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("25e25"))
	})

	It("preserves boolean-like strings without converting to bool", func() {
		r, err := op.Run(ev, []*Expr{
			env("RAW_ENV_BOOL"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("true"))
	})

	It("preserves numeric strings without converting to number", func() {
		r, err := op.Run(ev, []*Expr{
			env("RAW_ENV_NUM"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("42"))
	})

	It("preserves float strings without converting to number", func() {
		r, err := op.Run(ev, []*Expr{
			env("RAW_ENV_FLOAT"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("3.14159"))
	})

	It("preserves YAML list syntax as string without parsing", func() {
		r, err := op.Run(ev, []*Expr{
			env("RAW_ENV_YAML_LIST"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("[1, 2, 3]"))
	})

	It("preserves YAML map syntax as string without parsing", func() {
		r, err := op.Run(ev, []*Expr{
			env("RAW_ENV_YAML_MAP"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("{key: value}"))
	})

	It("preserves multiline strings", func() {
		r, err := op.Run(ev, []*Expr{
			env("RAW_ENV_MULTILINE"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("line1\n\nline3\nline4"))
	})

	It("allows empty string environment variables", func() {
		r, err := op.Run(ev, []*Expr{
			env("RAW_ENV_EMPTY_STRING"),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal(""))
	})

	It("falls back to second env var when first is unset", func() {
		r, err := op.Run(ev, []*Expr{
			or(env("RAW_ENV_UNSET"), env("RAW_ENV_TEST")),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("hello world"))
	})

	It("falls back to string literal when env var is unset", func() {
		r, err := op.Run(ev, []*Expr{
			or(env("RAW_ENV_UNSET"), str("default_value")),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("default_value"))
	})

	It("uses first env var when it is set (short-circuit)", func() {
		r, err := op.Run(ev, []*Expr{
			or(env("RAW_ENV_TEST"), str("fallback")),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("hello world"))
	})

	It("chained fallback preserves raw string", func() {
		r, err := op.Run(ev, []*Expr{
			or(env("RAW_ENV_UNSET"), or(env("RAW_ENV_ALSO_UNSET"), env("RAW_ENV_SCI"))),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("25e25"))
	})

	It("errors when all alternatives are unset", func() {
		_, err := op.Run(ev, []*Expr{
			or(env("RAW_ENV_UNSET"), env("RAW_ENV_ALSO_UNSET")),
		})
		Expect(err).To(HaveOccurred())
	})

	It("falls back to reference when env var is unset", func() {
		evWithTree := &Evaluator{
			Tree: opYAML("meta:\n  default: from_tree\n"),
		}
		r, err := op.Run(evWithTree, []*Expr{
			or(env("RAW_ENV_UNSET"), ref("meta.default")),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(r).NotTo(BeNil())
		Expect(r.Type).To(Equal(Replace))
		Expect(r.Value.(string)).To(Equal("from_tree"))
	})

	It("throws errors for unset environment variables", func() {
		_, err := op.Run(ev, []*Expr{
			env("RAW_ENV_UNSET"),
		})
		Expect(err).To(HaveOccurred())
	})

	It("throws errors for missing arguments", func() {
		_, err := op.Run(ev, []*Expr{})
		Expect(err).To(HaveOccurred())
	})

	It("throws errors when more than one argument provided", func() {
		_, err := op.Run(ev, []*Expr{
			env("RAW_ENV_TEST"),
			env("RAW_ENV_NUM"),
		})
		Expect(err).To(HaveOccurred())
	})

	It("throws errors when argument is not an environment variable", func() {
		_, err := op.Run(ev, []*Expr{
			str("not an env var"),
		})
		Expect(err).To(HaveOccurred())
	})

	It("throws errors when argument is a reference", func() {
		_, err := op.Run(ev, []*Expr{
			ref("some.reference"),
		})
		Expect(err).To(HaveOccurred())
	})
})
