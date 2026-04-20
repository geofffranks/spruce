# AWS SDK v1 → v2 Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace deprecated/EOL `github.com/aws/aws-sdk-go` v1 with `github.com/aws/aws-sdk-go-v2` for the `(( awsparam ))` / `(( awssecret ))` operators, using thin own-interfaces backed by v2 clients so counterfeiter fakes stay tiny.

**Architecture:** Define two single-method interfaces (`SSMClient`, `SecretsManagerClient`) in `op_aws.go`. Each matches the v2 client method signature exactly so `*ssm.Client` / `*secretsmanager.Client` satisfy them implicitly. Session-based credential loading becomes config-based (`config.LoadDefaultConfig` + in-place `cfg.Credentials = stscreds.NewAssumeRoleProvider(...)` for AWS_ROLE). Counterfeiter fakes regenerate against the new interfaces — each fake drops from ~40k lines to ~100. Tests swap v1 types/helpers for v2 equivalents; assertions stay semantically identical (caching, error messages, env-var handling unchanged).

**Tech Stack:** `github.com/aws/aws-sdk-go-v2` (core + config + credentials), `github.com/aws/aws-sdk-go-v2/service/ssm`, `github.com/aws/aws-sdk-go-v2/service/secretsmanager`, `github.com/aws/aws-sdk-go-v2/service/sts`, existing `github.com/maxbrunsfeld/counterfeiter/v6` (via `go tool`), existing Ginkgo/Gomega.

**Spec:** `docs/superpowers/specs/2026-04-17-aws-sdk-v2-migration-design.md`

**Local LLM delegation:** Steps marked [LLM-OK] are candidates for `local_llm` per the `local-llm-delegation` skill. See that skill for payload recipes. Read the file yourself, pass to `local_llm`, apply the result, verify by compile/test.

---

## File Structure

| File | Change | Responsibility |
|------|--------|----------------|
| `op_aws.go` | Rewrite | Operator impl + own interfaces + v2 config init |
| `operator_test.go` | Modify (lines 1-22 imports; 2035-2244 AWS tests) | Ginkgo tests, swap v1 types/helpers for v2 |
| `fakes/generate.go` | Rewrite | Counterfeiter directives targeting own interfaces |
| `fakes/fake_ssm_api.go` | Delete | Old v1 fake (superseded) |
| `fakes/fake_secrets_manager_api.go` | Delete | Old v1 fake (superseded) |
| `fakes/fake_ssm_client.go` | Create (generated) | New 1-method fake for `SSMClient` |
| `fakes/fake_secrets_manager_client.go` | Create (generated) | New 1-method fake for `SecretsManagerClient` |
| `go.mod`, `go.sum` | Modify | Drop v1, add v2 modules |
| `vendor/` | Modify | Drop v1 tree, add v2 tree via `go mod vendor` |

---

## Task 1: Add v2 SDK Modules, Keep v1 Temporarily

**Why first:** Establishes v2 deps before any source-code changes. Keeping v1 in parallel for this task means everything still compiles — subsequent tasks can stage incremental changes.

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`
- Modify: `vendor/` (entire tree)

- [ ] **Step 1: Add v2 modules**

Run:
```bash
go get github.com/aws/aws-sdk-go-v2@latest
go get github.com/aws/aws-sdk-go-v2/config@latest
go get github.com/aws/aws-sdk-go-v2/credentials@latest
go get github.com/aws/aws-sdk-go-v2/service/ssm@latest
go get github.com/aws/aws-sdk-go-v2/service/secretsmanager@latest
go get github.com/aws/aws-sdk-go-v2/service/sts@latest
```

Expected: `go.mod` and `go.sum` updated with new modules. No `require` entries removed yet.

- [ ] **Step 2: Revendor**

Run: `go mod vendor`

Expected: New `vendor/github.com/aws/aws-sdk-go-v2/` tree appears. Old `vendor/github.com/aws/aws-sdk-go/` tree remains (still used by op_aws.go).

- [ ] **Step 3: Verify build still passes**

Run: `go build ./...`

Expected: Build succeeds. No compilation errors (v1 still imported by op_aws.go; v2 imported nowhere yet).

- [ ] **Step 4: Verify tests still pass**

Run: `make test`

Expected: All tests pass. The SDK parallel state should not affect anything.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum vendor/
git commit -m "deps: add aws-sdk-go-v2 modules alongside v1

Pre-cursor to migration. Both SDKs coexist temporarily so subsequent
commits can swap op_aws.go and tests in isolation."
```

---

## Task 2: Rewrite `op_aws.go` Against v2 with Own Interfaces

**Why:** Core production swap. After this, `op_aws.go` no longer imports v1 — only v2 + own interfaces. Tests still reference v1 fakes, so they will FAIL to compile. Task 3 fixes that. We commit the broken intermediate on a feature branch since reviewers need to see these changes separately.

**Files:**
- Modify: `op_aws.go` (full rewrite of imports + vars + `initializeAwsSession` + `getAwsParam` + `getAwsSecret` + `Run`)

- [ ] **Step 1: Replace file contents**

Write the new `op_aws.go`:

```go
package spruce

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/starkandwayne/goutils/ansi"

	. "github.com/geofffranks/spruce/log"
	"github.com/starkandwayne/goutils/tree"

	// Use geofffranks forks to persist the fix in https://github.com/go-yaml/yaml/pull/133/commits
	// Also https://github.com/go-yaml/yaml/pull/195
	"github.com/geofffranks/yaml"
)

// SSMClient abstracts SSM Parameter Store access. The real v2 *ssm.Client
// satisfies this interface implicitly.
type SSMClient interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

// SecretsManagerClient abstracts Secrets Manager access. The real v2
// *secretsmanager.Client satisfies this interface implicitly.
type SecretsManagerClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// awsConfig holds a shared AWS config value
var awsConfig *aws.Config

// secretsManagerClient holds a secretsmanager client configured with a config.
// Test code replaces this with a counterfeiter fake.
var secretsManagerClient SecretsManagerClient

// parameterstoreClient holds a parameterstore client configured with a config.
// Test code replaces this with a counterfeiter fake.
var parameterstoreClient SSMClient

// awsSecretsCache caches values from AWS Secretsmanager
var awsSecretsCache = make(map[string]string)

// awsParamsCache caches values from AWS Parameterstore
var awsParamsCache = make(map[string]string)

// SkipAws toggles whether AwsOperator will attempt to query AWS for any value
// When true will always return "REDACTED"
var SkipAws bool

// AwsOperator provides two operators;  (( awsparam "path" )) and (( awssecret "name_or_arn" ))
// It will fetch parameters / secrets from the respective AWS service
type AwsOperator struct {
	variant string
}

// initializeAwsConfig builds an aws.Config honoring shared config (e.g. ~/.aws/credentials),
// optional profile, optional region, and optional STS AssumeRole.
func initializeAwsConfig(ctx context.Context, profile string, region string, role string) (*aws.Config, error) {
	var optFns []func(*config.LoadOptions) error
	if region != "" {
		optFns = append(optFns, config.WithRegion(region))
	}
	if profile != "" {
		optFns = append(optFns, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, err
	}

	if role != "" {
		stsClient := sts.NewFromConfig(cfg)
		// Wrap AssumeRoleProvider in CredentialsCache per v2 best practice —
		// without it, every service call could trigger a new STS AssumeRole.
		// config.LoadDefaultConfig sets up caching for the default chain, but
		// that wrapping is lost when we replace cfg.Credentials here.
		cfg.Credentials = aws.NewCredentialsCache(stscreds.NewAssumeRoleProvider(stsClient, role))
	}

	return &cfg, nil
}

// getAwsSecret will fetch the specified secret from AWS Secretsmanager at the specified (if provided) stage / version
func getAwsSecret(ctx context.Context, cfg *aws.Config, secret string, params url.Values) (string, error) {
	val, cached := awsSecretsCache[secret]
	if cached {
		return val, nil
	}

	if secretsManagerClient == nil {
		secretsManagerClient = secretsmanager.NewFromConfig(*cfg)
	}

	input := secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secret),
	}

	if params.Get("stage") != "" {
		input.VersionStage = aws.String(params.Get("stage"))
	} else if params.Get("version") != "" {
		input.VersionId = aws.String(params.Get("version"))
	}

	output, err := secretsManagerClient.GetSecretValue(ctx, &input)
	if err != nil {
		return "", err
	}

	awsSecretsCache[secret] = aws.ToString(output.SecretString)

	return awsSecretsCache[secret], nil
}

// getAwsParam will fetch the specified parameter from AWS SSM Parameterstore
func getAwsParam(ctx context.Context, cfg *aws.Config, param string) (string, error) {
	val, cached := awsParamsCache[param]
	if cached {
		return val, nil
	}

	if parameterstoreClient == nil {
		parameterstoreClient = ssm.NewFromConfig(*cfg)
	}

	input := ssm.GetParameterInput{
		Name:           aws.String(param),
		WithDecryption: aws.Bool(true),
	}

	output, err := parameterstoreClient.GetParameter(ctx, &input)
	if err != nil {
		return "", err
	}

	awsParamsCache[param] = aws.ToString(output.Parameter.Value)

	return awsParamsCache[param], nil
}

// Setup ...
func (AwsOperator) Setup() error {
	return nil
}

// Phase ...
func (AwsOperator) Phase() OperatorPhase {
	return EvalPhase
}

// Dependencies is not used by AwsOperator
func (AwsOperator) Dependencies(_ *Evaluator, _ []*Expr, _ []*tree.Cursor, auto []*tree.Cursor) []*tree.Cursor {
	return auto
}

// Run will invoke the appropriate getAws* function for each instance of the AwsOperator
// and extract the specified key (if provided).
func (o AwsOperator) Run(ev *Evaluator, args []*Expr) (*Response, error) {
	var err error
	DEBUG("running (( %s ... )) operation at $.%s", o.variant, ev.Here)
	defer DEBUG("done with (( %s ... )) operation at $.%s\n", o.variant, ev.Here)

	if len(args) < 1 {
		return nil, fmt.Errorf("%s operator requires at least one argument", o.variant)
	}

	var l []string
	for i, arg := range args {
		v, err := arg.Resolve(ev.Tree)
		if err != nil {
			DEBUG("  arg[%d]: failed to resolve expression to a concrete value", i)
			DEBUG("     [%d]: error was: %s", i, err)
			return nil, err
		}

		switch v.Type {
		case Literal:
			DEBUG("  arg[%d]: using string literal '%v'", i, v.Literal)
			l = append(l, fmt.Sprintf("%v", v.Literal))

		case Reference:
			DEBUG("  arg[%d]: trying to resolve reference $.%s", i, v.Reference)
			s, err := v.Reference.Resolve(ev.Tree)
			if err != nil {
				DEBUG("     [%d]: resolution failed\n    error: %s", i, err)
				return nil, fmt.Errorf("unable to resolve `%s`: %s", v.Reference, err)
			}

			switch s.(type) {
			case map[interface{}]interface{}:
				DEBUG("  arg[%d]: %v is not a string scalar", i, s)
				return nil, ansi.Errorf("@c{$.%s}@R{ is a map; only scalars are supported here}", v.Reference)

			case []interface{}:
				DEBUG("  arg[%d]: %v is not a string scalar", i, s)
				return nil, ansi.Errorf("@c{$.%s}@R{ is a list; only scalars are supported here}", v.Reference)

			default:
				l = append(l, fmt.Sprintf("%v", s))
			}

		default:
			DEBUG("  arg[%d]: I don't know what to do with '%v'", i, arg)
			return nil, fmt.Errorf("%s operator only accepts string literals and key reference arguments", o.variant)
		}
	}

	key, params, err := parseAwsOpKey(strings.Join(l, ""))
	if err != nil {
		return nil, err
	}

	DEBUG("     [0]: Using %s key '%s'\n", o.variant, key)

	value := "REDACTED"

	if !SkipAws {
		ctx := context.Background()
		if awsConfig == nil {
			awsConfig, err = initializeAwsConfig(ctx, os.Getenv("AWS_PROFILE"), os.Getenv("AWS_REGION"), os.Getenv("AWS_ROLE"))
			if err != nil {
				return nil, fmt.Errorf("error during AWS config initialization: %s", err)
			}
		}

		if o.variant == "awsparam" {
			value, err = getAwsParam(ctx, awsConfig, key)
		} else if o.variant == "awssecret" {
			value, err = getAwsSecret(ctx, awsConfig, key, params)
		}

		if err != nil {
			return nil, fmt.Errorf("$.%s error fetching %s: %s", key, o.variant, err)
		}

		subkey := params.Get("key")
		if subkey != "" {
			tmp := make(map[string]interface{})
			err := yaml.Unmarshal([]byte(value), &tmp)

			if err != nil {
				return nil, fmt.Errorf("$.%s error extracting key: %s", key, err)
			}

			if _, ok := tmp[subkey]; !ok {
				return nil, fmt.Errorf("$.%s invalid key '%s'", key, subkey)
			}

			value = fmt.Sprintf("%v", tmp[subkey])
		}
	}

	return &Response{
		Type:  Replace,
		Value: value,
	}, nil
}

// parseAwsOpKey parsed the parameters passed to AwsOperator.
// Primarily it splits the key from the extra arguments (specified as a query string)
func parseAwsOpKey(key string) (string, url.Values, error) {
	split := strings.SplitN(key, "?", 2)
	if len(split) == 1 {
		split = append(split, "")
	}

	values, err := url.ParseQuery(split[1])
	if err != nil {
		return "", values, fmt.Errorf("invalid argument string: %s", err)
	}

	return split[0], values, nil
}

// init registers the two variants of the AwsOperator
func init() {
	RegisterOp("awsparam", AwsOperator{variant: "awsparam"})
	RegisterOp("awssecret", AwsOperator{variant: "awssecret"})
}
```

Key differences from the v1 version:
- Error message: `"error during AWS session initialization"` → `"error during AWS config initialization"`. This error only surfaces if `config.LoadDefaultConfig` fails (bad shared config) — it's not covered by any existing test, so no test assertion to update.
- Package var renamed: `awsSession *session.Session` → `awsConfig *aws.Config`.
- `getAwsParam` / `getAwsSecret` gain `ctx context.Context` as first arg; `awsSession *session.Session` param becomes `cfg *aws.Config`.

- [ ] **Step 2: Build `op_aws.go` alone (expect test failures, production passes)**

Run: `go build ./...`

Expected: BUILD FAILS. Test compilation fails with errors about `fakes.FakeSSMAPI` not matching `SSMClient`, and unresolved `ssmiface`/`secretsmanageriface`. This is fine — Task 3/4 fix the fakes and tests.

If production (non-test) build fails, stop and debug — only test compilation should break here.

- [ ] **Step 3: Commit the broken-test intermediate**

```bash
git add op_aws.go
git commit -m "refactor: rewrite op_aws.go against aws-sdk-go-v2

Introduces thin SSMClient and SecretsManagerClient interfaces satisfied
implicitly by v2 service clients. Replaces session-based init with
config.LoadDefaultConfig + in-place AssumeRoleProvider for AWS_ROLE.

Operator behavior unchanged: cache, SkipAws redaction, env vars, and
query-string params (?key=, ?stage=, ?version=) all preserved.

Tests will fail to compile until counterfeiter fakes regenerate against
the new interfaces (next commit)."
```

---

## Task 3: Regenerate Counterfeiter Fakes Against Own Interfaces

**Why:** With `op_aws.go` using own interfaces, the v1-based fakes are obsolete. New fakes will be one stub per interface method — each interface has one method, so each fake is ~100 lines instead of ~40,000.

**Files:**
- Rewrite: `fakes/generate.go`
- Delete: `fakes/fake_ssm_api.go`
- Delete: `fakes/fake_secrets_manager_api.go`
- Create: `fakes/fake_ssm_client.go` (generated)
- Create: `fakes/fake_secrets_manager_client.go` (generated)

- [ ] **Step 1: Rewrite `fakes/generate.go`**

Replace file contents with:

```go
package fakes

//go:generate go tool counterfeiter -o fake_ssm_client.go github.com/geofffranks/spruce.SSMClient
//go:generate go tool counterfeiter -o fake_secrets_manager_client.go github.com/geofffranks/spruce.SecretsManagerClient
```

- [ ] **Step 2: Delete old fakes**

Run:
```bash
rm fakes/fake_ssm_api.go fakes/fake_secrets_manager_api.go
```

- [ ] **Step 3: Generate new fakes**

Run: `go generate ./fakes/...`

Expected: Creates `fakes/fake_ssm_client.go` and `fakes/fake_secrets_manager_client.go`. Each file should be ~80-120 lines with types `FakeSSMClient` / `FakeSecretsManagerClient` having `GetParameterStub` / `GetSecretValueStub` hooks, `GetParameterReturns` / `GetSecretValueReturns` helpers, and `CallCount` / `ArgsForCall` introspection helpers.

- [ ] **Step 4: Verify fakes build**

Run: `go build ./fakes/...`

Expected: `fakes/` package compiles clean. If counterfeiter emitted bad output (e.g., unresolved imports), re-run `go generate` after confirming `op_aws.go` compiles.

- [ ] **Step 5: Commit fakes regeneration**

```bash
git add fakes/
git commit -m "test: regenerate counterfeiter fakes against own SSM/SecretsManager interfaces

Old fakes were generated against v1 ssmiface.SSMAPI /
secretsmanageriface.SecretsManagerAPI — each ~40k lines covering the
full SDK surface. New fakes target single-method SSMClient /
SecretsManagerClient interfaces defined in op_aws.go — ~100 lines each.

Tests still fail to compile until operator_test.go is updated to
reference the new fake type names (next commit)."
```

---

## Task 4: Migrate `operator_test.go` AWS Tests to v2

**Why:** Final piece — tests reference the new fake types (`FakeSSMClient` / `FakeSecretsManagerClient`), v2 input/output types, v2 stub signatures with `context.Context`, and v2 helper `aws.ToString`. [LLM-OK for the mechanical substitution within the `awsparam/awssecret operator` Describe block.]

**Files:**
- Modify: `operator_test.go`
  - Lines 10-12: v1 imports (`aws`, `service/secretsmanager`, `service/ssm`)
  - Lines 2035-2244: AWS test Describe block

- [ ] **Step 1: Update imports at top of file**

Replace lines 10-12 (v1 imports) with v2 equivalents. The exact edit:

Old:
```go
	"github.com/aws/aws-sdk-go/aws"                    //nolint:staticcheck // SA1019: aws-sdk-go v1 deprecated; v2 migration tracked separately
	"github.com/aws/aws-sdk-go/service/secretsmanager" //nolint:staticcheck // SA1019: aws-sdk-go v1 deprecated; v2 migration tracked separately
	"github.com/aws/aws-sdk-go/service/ssm"            //nolint:staticcheck // SA1019: aws-sdk-go v1 deprecated; v2 migration tracked separately
```

New:
```go
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
```

No other imports in this file touch the AWS SDK (the `fakes` package import stays the same; it refers to `fakes.FakeSSMAPI` and `fakes.FakeSecretsManagerAPI` in code, which Step 2 updates).

- [ ] **Step 2: Rewrite AWS Describe block** [LLM-OK]

Replace lines 2035-2244 with the block below. All changes are mechanical:
- `fakes.FakeSSMAPI` → `fakes.FakeSSMClient`
- `fakes.FakeSecretsManagerAPI` → `fakes.FakeSecretsManagerClient`
- Stub signature `func(in *ssm.GetParameterInput)` → `func(ctx context.Context, in *ssm.GetParameterInput, optFns ...func(*ssm.Options))`
- Stub signature `func(in *secretsmanager.GetSecretValueInput)` → `func(ctx context.Context, in *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options))`
- Input struct types unchanged (same `ssm.GetParameterInput`, `secretsmanager.GetSecretValueInput`, `ssm.Parameter`, etc. — they exist at the same names in v2)
- `aws.StringValue(x)` → `aws.ToString(x)`
- `aws.String(...)` unchanged (exists in both v1 and v2)
- `aws.Bool(...)` unchanged (not used in this block, but would also be unchanged)
- Add `"context"` to the imports at the top if not already present (it is: line 4)

New block (replaces lines 2035-2244):

```go
var _ = Describe("awsparam/awssecret operator", func() {
	var op AwsOperator
	var ev *Evaluator
	var fakeSSM *fakes.FakeSSMClient
	var fakeSecretsManager *fakes.FakeSecretsManagerClient

	BeforeEach(func() {
		op = AwsOperator{variant: "awsparam"}
		ev = &Evaluator{
			Tree: opYAML(`{ "testval": "test", "testmap": {}, "testarr": [] }`),
			Here: &tree.Cursor{},
		}
		fakeSSM = new(fakes.FakeSSMClient)
		fakeSecretsManager = new(fakes.FakeSecretsManagerClient)
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
			fakeSSM.GetParameterStub = func(ctx context.Context, in *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
				ssmKey = aws.ToString(in.Name)
				return &ssm.GetParameterOutput{
					Parameter: &ssmtypes.Parameter{
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
			fakeSSM.GetParameterStub = func(ctx context.Context, in *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
				ssmKey = aws.ToString(in.Name)
				return &ssm.GetParameterOutput{
					Parameter: &ssmtypes.Parameter{
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
			fakeSSM.GetParameterStub = func(ctx context.Context, in *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
				return &ssm.GetParameterOutput{
					Parameter: &ssmtypes.Parameter{
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
				fakeSSM.GetParameterStub = func(ctx context.Context, in *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
					return &ssm.GetParameterOutput{
						Parameter: &ssmtypes.Parameter{
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
				fakeSSM.GetParameterStub = func(ctx context.Context, in *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
					return &ssm.GetParameterOutput{
						Parameter: &ssmtypes.Parameter{
							Value: aws.String(`key: {`),
						},
					}, nil
				}
				_, err := op.Run(ev, []*Expr{str("val3?key=key")})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("$.val3 error extracting key: yaml: line 1: did not find expected node content"))
			})

			It("should error if subkey invalid", func() {
				fakeSSM.GetParameterStub = func(ctx context.Context, in *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
					return &ssm.GetParameterOutput{
						Parameter: &ssmtypes.Parameter{
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
			fakeSSM.GetParameterStub = func(ctx context.Context, in *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
				count++
				return &ssm.GetParameterOutput{
					Parameter: &ssmtypes.Parameter{
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
			fakeSSM.GetParameterStub = func(ctx context.Context, in *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
				count++
				return &ssm.GetParameterOutput{
					Parameter: &ssmtypes.Parameter{
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
			fakeSecretsManager.GetSecretValueStub = func(ctx context.Context, in *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
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
			fakeSecretsManager.GetSecretValueStub = func(ctx context.Context, in *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				stage = aws.ToString(in.VersionStage)
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
			fakeSecretsManager.GetSecretValueStub = func(ctx context.Context, in *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				version = aws.ToString(in.VersionId)
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
```

**Note on `ssmtypes.Parameter`:** In v2, `ssm.Parameter` moved to a separate types package at `github.com/aws/aws-sdk-go-v2/service/ssm/types`. The block above references it as `ssmtypes.Parameter`, so Step 3 adds the import. `secretsmanager.GetSecretValueOutput` keeps `SecretString` on the output struct itself (no types-package move).

- [ ] **Step 3: Add ssmtypes import**

Add this import to `operator_test.go` (group with other AWS imports at the top):

```go
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
```

- [ ] **Step 4: Run tests**

Run: `go tool ginkgo -r --race --fail-on-pending --keep-going --fail-on-empty --require-suite ./...`

Expected: All tests pass, including the 12 AWS test cases:
- 1x "should return error if no key given"
- 1x "should concatenate args"
- 1x "should resolve references"
- 1x "should not allow references to maps"
- 1x "should not allow references to arrays"
- 1x "without key"
- 3x "with key" (subkey extract / invalid yaml / invalid subkey)
- 1x "should not call AWS API if SkipAws true"
- 1x "awsparam should cache lookups"
- 3x "awssecret" (cache / stage / version)

Total: 12 tests in the AWS Describe block. If any fail, the most likely causes:
- Missing `ssmtypes` import → `undefined: ssmtypes` compile error
- Stub signature mismatch → counterfeiter complains at compile time about assignment to `GetParameterStub`
- Global state leak from prior tests setting `awsSession`/`awsConfig` → check `BeforeEach` clears `awsConfig`, `awsSecretsCache`, `awsParamsCache` if this becomes an issue (existing code doesn't, so likely fine).

- [ ] **Step 5: Commit**

```bash
git add operator_test.go
git commit -m "test: migrate AWS operator tests to aws-sdk-go-v2

Update counterfeiter fake types (FakeSSMAPI → FakeSSMClient,
FakeSecretsManagerAPI → FakeSecretsManagerClient), stub signatures to
match v2 method shape (context.Context + variadic options), and swap
aws.StringValue calls for aws.ToString. ssm.Parameter reference moves
to ssmtypes.Parameter (v2 moved it to the types subpackage).

Assertions and test behavior unchanged."
```

---

## Task 5: Drop v1 SDK and Revendor

**Why:** With all call sites migrated, the v1 module is unreferenced. Drop it from go.mod, tidy, revendor. This reclaims vendor disk space (~2MB+ of v1 tree) and makes the deprecation lint clean.

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`
- Modify: `vendor/` (removes `vendor/github.com/aws/aws-sdk-go/`)

- [ ] **Step 1: Tidy modules**

Run: `go mod tidy`

Expected: `github.com/aws/aws-sdk-go v1.55.8` and `// indirect` entries it pulled in drop out of `go.mod`. `go.sum` entries for v1 drop out. V2 entries remain. If any v1 entry remains after `go mod tidy`, grep the codebase to find the remaining import.

- [ ] **Step 2: Revendor**

Run: `go mod vendor`

Expected: `vendor/github.com/aws/aws-sdk-go/` tree removed. `vendor/modules.txt` updated accordingly.

- [ ] **Step 3: Sanity-check no v1 references remain**

Run:
```bash
rtk grep "github.com/aws/aws-sdk-go\\b" --include="*.go" -r
```

Expected: Zero matches outside `vendor/` (after revendor there should be zero total). If any non-vendor match shows up, it's a missed import — fix before proceeding.

- [ ] **Step 4: Full build + test + lint**

Run: `make all`

Expected: `vet`, `lint` (gofmt + staticcheck + gosec), `test` (Ginkgo), and `build` all pass. The `SA1019: aws-sdk-go` deprecation warnings disappear entirely.

If staticcheck still warns about v1, a reference is lurking somewhere — re-run the grep in Step 3.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum vendor/
git commit -m "deps: drop aws-sdk-go v1, purge vendored tree

All call sites now use aws-sdk-go-v2. go mod tidy drops v1 and its
transitive indirect entries; go mod vendor removes the vendored v1
tree (~2MB). staticcheck SA1019 deprecation warnings for aws-sdk-go
are now clean."
```

---

## Task 6: Final Verification

- [ ] **Step 1: Clean tree check**

Run: `git status`

Expected: Working tree clean. All changes committed.

- [ ] **Step 2: Grep for v1 residue**

Run:
```bash
rtk grep "aws-sdk-go\\b" --include="*.go" -r
rtk grep "SA1019" --include="*.go" -r
rtk grep "ssmiface\\|secretsmanageriface" --include="*.go" -r
rtk grep "aws\\.StringValue" --include="*.go" -r
rtk grep "session\\.NewSessionWithOptions\\|session\\.NewSession" --include="*.go" -r
```

Expected: Zero matches for any. (The `SA1019` grep should also show nothing — we've removed every nolint suppression tied to aws-sdk-go.)

- [ ] **Step 3: Full Makefile flow**

Run: `make all`

Expected: Clean pass on vet, lint, test, build. Race detector enabled in ginkgo target — the test run exercises fake concurrency.

- [ ] **Step 4: Smoke test the operator end-to-end (optional, local AWS creds required)**

If you have local AWS creds and a sandbox Parameter Store entry, run a quick manual validation:

```bash
echo 'value: (( awsparam "/spruce-test/dummy" ))' | ./spruce merge
```

Expected: Either the real parameter value prints, or a realistic AWS API error (missing param, permission denied, etc.). Do NOT expect `REDACTED` unless `SkipAws` is programmatically set — that's a library-level toggle.

If no AWS creds available, skip this step — the ginkgo tests already cover all code paths through fakes.

- [ ] **Step 5: (Optional) Announce completion via finishing-a-development-branch**

The feature branch is ready for merge/PR. Invoke `finishing-a-development-branch` skill (user-local variant) to delegate PR title/body drafting to `local_llm` and get merge/PR recommendations.

---

## Self-Review Notes (resolved during planning)

- **Spec coverage:**
  - Section 1 (Own Interfaces) → Task 2 Step 1 body defines `SSMClient` + `SecretsManagerClient`
  - Section 2 (Session → Config) → Task 2 Step 1 body replaces `initializeAwsSession` with `initializeAwsConfig`; STS AssumeRole pattern matches spec exactly
  - Section 3 (Client Construction + DI) → Task 2 Step 1 body uses `ssm.NewFromConfig` / `secretsmanager.NewFromConfig`; `ctx context.Context` added to getAwsParam/getAwsSecret; `Run()` passes `context.Background()`
  - Section 4 (Type/Import Changes) → Task 2 import block + Task 4 import edits; Task 5 Step 3 grep verifies zero v1 residue
  - Section 5 (Counterfeiter Fake Regeneration) → Task 3 entirely; uses `go tool counterfeiter` per spec update
  - Section 6 (Cleanup) → Task 5 (mod tidy, vendor, nolint removal); Task 6 (verification)
- **Placeholder scan:** No "TBD"/"TODO"/"implement later" patterns. Every `//nolint` comment is explicitly removed; the two `BeforeEach` error-state comments in the original code are also unchanged (they describe real behavior, not plan placeholders).
- **Type consistency:** `SSMClient` / `SecretsManagerClient` interface names consistent across op_aws.go (Task 2), generate.go (Task 3 Step 1), operator_test.go (Task 4 Step 2). Fake type names `FakeSSMClient` / `FakeSecretsManagerClient` match counterfeiter output convention. `awsConfig` is consistently a `*aws.Config` across all references.
- **Branch coverage:** All 12 existing Ginkgo AWS test cases preserved; all 4 error branches in `op_aws.go` (`config.LoadDefaultConfig` error, `getAwsParam` error, `getAwsSecret` error, and invalid-subkey error) — same coverage as pre-migration since we aren't adding or removing behavior.
- **No condition-form/truth-table rationale needed:** This plan has no condition-form choices (positive vs. negative matches, native vs. template conditions). Interface and implementation choices are structural, not state-space-dependent.

---

## Execution Tips

- **Run tasks 1 → 5 in order.** Task 2 leaves tests broken; Task 3 leaves them still broken; Task 4 restores green. Don't split this work across multiple PRs — the intermediate commits don't compile on their own and shouldn't land individually.
- **If `go generate` fails in Task 3:** counterfeiter needs the interface's package to compile. Confirm `go build .` works on the project root before re-running generate.
- **If a test fails in Task 4:** diff the stub signatures against the counterfeiter-generated fake's `GetParameterStub` type — they must match exactly. Mismatches here fail at compile time with a clear error.
- **Race detector:** the Makefile's ginkgo target already passes `--race`. The counterfeiter fakes are thread-safe out of the box, but if a race surfaces inside op_aws.go caches (unrelated to this migration — caches were never locked), note it for a follow-up issue; do not fix in this PR.
