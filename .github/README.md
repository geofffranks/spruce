# CI/CD Overview

## Workflows

### `test.yml` — PR Checks

Runs on every push and pull request. Single job on `ubuntu-latest`:

1. **Build** — `go build -v ./...`
2. **Test** — `go test` across all non-vendor packages
3. **Vet** — `go vet` across all non-vendor packages
4. **staticcheck** — static analysis via `dominikh/staticcheck-action`
5. **gosec** — security scanner via `securego/gosec`

Go version is `stable` with `check-latest: true`, so CI always uses the latest released Go.

### `release.yml` — Release Artifacts

Triggers when a GitHub release is created. Builds cross-platform binaries (linux/amd64, linux/arm64, windows/amd64, darwin/amd64, darwin/arm64) with the release tag embedded via `-ldflags`, then uploads each binary and its SHA1 checksum as release assets.

Uses `go-version: stable` — artifacts are always built with the latest Go at release time.

### `codeql-analysis.yml` — Security Scanning

Runs CodeQL analysis for Go on pushes and PRs to `main`, plus a weekly schedule (Tuesdays 23:40 UTC). Catches security vulnerabilities and coding errors.

### `dependabot-auto-merge.yml` — Auto-Merge Dependabot PRs

Triggers on `pull_request_target` for PRs opened by `dependabot[bot]`. Uses `dependabot/fetch-metadata` to classify the version bump, then enables auto-merge (rebase strategy) for **patch and minor** updates only. Major version bumps are left for manual review.

Auto-merge is gated by branch protection — the PR won't actually merge until the `test` and `Analyze (go)` status checks pass.

> **Security note:** This workflow deliberately does not check out PR code. It only reads dependabot metadata and calls `gh pr merge`. A comment at the top of the file explains why — do not add a checkout step without understanding the `pull_request_target` security model.

### `go-version-bump.yml` — Go Toolchain Auto-Bump

Runs daily at 14:00 UTC (plus manual `workflow_dispatch`). Checks `go.dev/dl/?mode=json` for the latest stable Go release and compares it to the `toolchain` directive in `go.mod`. If they differ:

1. Creates branch `auto/bump-go-toolchain-<version>`
2. Runs `go mod edit -toolchain=go<version>` + `go mod tidy`
3. Opens a PR and enables auto-merge (rebase)

Idempotent — if the branch already exists from a prior run, the workflow exits cleanly. The `go X.Y` minimum-version directive in `go.mod` is intentionally left alone; only the `toolchain` line floats.

## Keeping Things Fresh

| What | How | Frequency |
|---|---|---|
| Go modules | Dependabot | Weekly PRs, patch/minor auto-merged |
| GitHub Action pins | Dependabot | Weekly PRs, patch/minor auto-merged |
| Go toolchain in `go.mod` | `go-version-bump.yml` | Daily check, PR when new stable ships |
| Go in CI runners | `go-version: stable` | Always latest (no pin to manage) |

## Branch Protection

`main` requires the following checks before merge:

- **`test`** — build, test, vet, staticcheck, gosec
- **`Analyze (go)`** — CodeQL security analysis

Auto-merge is enabled. PRs queued with `gh pr merge --auto --rebase` will merge only after both checks pass.

## Configuration Files

- [`dependabot.yml`](dependabot.yml) — Dependabot config for `gomod` and `github-actions` ecosystems, targeting `main`, weekly cadence
- [`workflows/test.yml`](workflows/test.yml) — PR checks
- [`workflows/release.yml`](workflows/release.yml) — Release artifact builder
- [`workflows/codeql-analysis.yml`](workflows/codeql-analysis.yml) — CodeQL security scanning
- [`workflows/dependabot-auto-merge.yml`](workflows/dependabot-auto-merge.yml) — Dependabot auto-merge logic
- [`workflows/go-version-bump.yml`](workflows/go-version-bump.yml) — Go toolchain auto-bump
