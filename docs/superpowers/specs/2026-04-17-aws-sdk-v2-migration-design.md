# AWS SDK v1 → v2 Migration Design

**Date:** 2026-04-17 (updated 2026-04-20)
**Prior work:** Ginkgo test migration (`2026-04-16-ginkgo-test-migration.md`) is complete. Tests run under Ginkgo/Gomega; counterfeiter fakes live in `fakes/` and target v1 `ssmiface.SSMAPI` / `secretsmanageriface.SecretsManagerAPI`. This plan replaces those with own interfaces backed by v2 clients.

## Problem

`github.com/aws/aws-sdk-go` v1 (v1.55.8) is deprecated and EOL. Spruce depends on it for the `(( awsparam ))` and `(( awssecret ))` operators in `op_aws.go`. The SDK needs to migrate to `github.com/aws/aws-sdk-go-v2`.

## Scope

Files affected:

| File | Role |
|------|------|
| `op_aws.go` | Production code (~200 lines): session init, SSM GetParameter, SecretsManager GetSecretValue, STS AssumeRole. Uses `ssmiface.SSMAPI` / `secretsmanageriface.SecretsManagerAPI` typed vars. |
| `operator_test.go` | Ginkgo tests referencing `fakes.FakeSSMAPI` / `fakes.FakeSecretsManagerAPI` (4 refs). |
| `fakes/generate.go` | Counterfeiter directives targeting v1 `ssmiface`/`secretsmanageriface`. Uses `go tool counterfeiter` (go.mod tool directive). |
| `fakes/fake_ssm_api.go`, `fakes/fake_secrets_manager_api.go` | Existing generated fakes (large: full SDK interface). Delete + regenerate against own interfaces. |
| `go.mod`, `go.sum`, `vendor/` | Drop v1 modules, add v2 modules, revendor. |

## Approach

Approach C from brainstorming: ginkgo migration already done (fakes against v1 `*iface`). This plan swaps to v2 with own interfaces and regenerates fakes. The regenerated fakes will be dramatically smaller — counterfeiter generates one stub per interface method, and the new interfaces define 1 method each vs. the ~300 methods in v1 `ssmiface.SSMAPI`.

## Design

### 1. Own Interfaces

Define two thin interfaces in `op_aws.go` covering only the methods actually called:

```go
type SSMClient interface {
    GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

type SecretsManagerClient interface {
    GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}
```

These match the v2 client method signatures exactly — `ssm.Client` and `secretsmanager.Client` satisfy them implicitly with no wrapper needed.

Package-level vars change type:
```go
var parameterstoreClient SSMClient              // was ssmiface.SSMAPI
var secretsManagerClient SecretsManagerClient   // was secretsmanageriface.SecretsManagerAPI
```

### 2. Session → Config Migration

`initializeAwsSession` becomes `initializeAwsConfig` returning `aws.Config`:

| v1 | v2 |
|----|-----|
| `session.NewSessionWithOptions(opts)` | `config.LoadDefaultConfig(ctx, optFns...)` |
| `SharedConfigState: SharedConfigEnable` | Default behavior (always loads shared config) |
| `options.Config.Region = aws.String(region)` | `config.WithRegion(region)` |
| `options.Profile = profile` | `config.WithSharedConfigProfile(profile)` |
| `stscreds.NewCredentials(s, role)` | See below |

**STS AssumeRole pattern in v2:**
```go
// First load base config
cfg, err := config.LoadDefaultConfig(ctx, optFns...)

// If role is set, create STS client from base config, then wrap with AssumeRoleProvider
if role != "" {
    stsClient := sts.NewFromConfig(cfg)
    cfg.Credentials = stscreds.NewAssumeRoleProvider(stsClient, role)
}
```
v1 required creating a second session; v2 mutates the config's credential provider in-place.

Package var: `awsSession *session.Session` → `awsConfig *aws.Config`

### 3. Client Construction + Dependency Injection

Same lazy-init pattern, just v2 constructors:

```go
if parameterstoreClient == nil {
    parameterstoreClient = ssm.NewFromConfig(*awsConfig)
}
```

`getAwsParam` and `getAwsSecret` signatures gain `context.Context`. `Run()` passes `context.Background()`.

DI for tests unchanged — counterfeiter fakes injected into package vars in `BeforeEach`, as established by the ginkgo migration. Fake variable names in `operator_test.go` change: `fakeSSM *fakes.FakeSSMAPI` → `fakeSSM *fakes.FakeSSMClient`; same for SecretsManager.

### 4. Type/Import Changes

| v1 import | v2 import |
|-----------|-----------|
| `github.com/aws/aws-sdk-go/aws` | `github.com/aws/aws-sdk-go-v2/aws` |
| `github.com/aws/aws-sdk-go/aws/session` | `github.com/aws/aws-sdk-go-v2/config` |
| `github.com/aws/aws-sdk-go/aws/credentials/stscreds` | `github.com/aws/aws-sdk-go-v2/credentials/stscreds` |
| `github.com/aws/aws-sdk-go/service/ssm` | `github.com/aws/aws-sdk-go-v2/service/ssm` |
| `github.com/aws/aws-sdk-go/service/ssm/ssmiface` | *(removed — own interface)* |
| `github.com/aws/aws-sdk-go/service/secretsmanager` | `github.com/aws/aws-sdk-go-v2/service/secretsmanager` |
| `github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface` | *(removed — own interface)* |
| *(none)* | `github.com/aws/aws-sdk-go-v2/service/sts` |

Helper changes:
- `aws.StringValue()` → `aws.ToString()`
- `aws.String()` — unchanged (exists in v2)
- `aws.Bool()` — unchanged (exists in v2)

### 5. Counterfeiter Fake Regeneration

`fakes/generate.go` directives update. Match the existing `go tool counterfeiter` pattern (counterfeiter is registered as a tool in `go.mod`, not run via `go run`):

```go
//go:generate go tool counterfeiter -o fake_ssm_client.go github.com/geofffranks/spruce.SSMClient
//go:generate go tool counterfeiter -o fake_secrets_manager_client.go github.com/geofffranks/spruce.SecretsManagerClient
```

Old fakes (`fake_ssm_api.go`, `fake_secrets_manager_api.go`) deleted. New fakes will be ~1/300th the size (one stub per method, and each new interface has exactly one method).

Test updates in `operator_test.go`:
- `fakes.FakeSSMAPI` → `fakes.FakeSSMClient`
- `fakes.FakeSecretsManagerAPI` → `fakes.FakeSecretsManagerClient`
- Counterfeiter helper names preserve semantics: `GetParameterReturns`, `GetParameterStub`, `GetParameterCallCount`, `GetParameterArgsForCall` — but stub signatures now include `context.Context` as first arg and `...func(*ssm.Options)` as trailing variadic.
- Input/output struct types change from v1 to v2 package paths.

### 6. Cleanup

- Remove `github.com/aws/aws-sdk-go v1.55.8` from `go.mod`
- `go mod tidy` + `go mod vendor`
- Remove all `//nolint:staticcheck // SA1019: aws-sdk-go v1 deprecated` comments
- `make test` + `make lint` — verify clean

**go.mod adds:**
- `github.com/aws/aws-sdk-go-v2`
- `github.com/aws/aws-sdk-go-v2/config`
- `github.com/aws/aws-sdk-go-v2/credentials`
- `github.com/aws/aws-sdk-go-v2/service/ssm`
- `github.com/aws/aws-sdk-go-v2/service/secretsmanager`
- `github.com/aws/aws-sdk-go-v2/service/sts`

## Subagent Model Guidance

For tasks involving mechanical code transformation, use the **`local_llm` MCP tool** (LM Studio, currently configured for Gemma 4 e4b 8bit per CLAUDE.md hardware tuning). Read the file content yourself, pass it to `local_llm` with clear instructions, then apply the result. The `local-llm-delegation` skill describes when and how.

**Good for `local_llm`:**
- **Import rewriting** — mechanical v1→v2 import path swaps in `op_aws.go` and `operator_test.go`
- **`aws.StringValue` → `aws.ToString`** — bulk replacement with context validation
- **Nolint comment removal** — identify and strip `//nolint:staticcheck // SA1019` suppressions, verify no unrelated suppressions lost
- **Counterfeiter fake spot-check** — pass generated fake code to verify it matches the new interface signatures
- **Test assertion updates** — adapting counterfeiter fake usage from `FakeSSMAPI` to `FakeSSMClient` and from v1 to v2 input/output struct types
- **Commit message drafts** — per granular-commits memory, one commit per logical change category

**Not suitable for `local_llm`:**
- `stscreds` migration (credential provider chain logic requires understanding v2 config loading)
- `initializeAwsConfig` rewrite (v2 config option composition semantics)
- Debugging build/test failures
- Architectural decisions about interface design

## Behavioral Compatibility

No user-visible behavior changes. The `(( awsparam ))` and `(( awssecret ))` operators continue to work identically. Environment variables (`AWS_PROFILE`, `AWS_REGION`, `AWS_ROLE`) and query-string parameters (`?key=`, `?stage=`, `?version=`) are unchanged. Caching behavior is preserved. `SkipAws` flag works the same.
