all: vet lint test build

vet:
	go vet ./...

lint: vet
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*' -not -path './.worktrees/*'))" || (echo "gofmt check failed:"; gofmt -l $$(find . -name '*.go' -not -path './vendor/*' -not -path './.worktrees/*'); exit 1)
	go tool staticcheck ./...
	go tool gosec ./...

# -p is intentionally omitted: vault_test.go uses os.Setenv globally and is
# not safe to run concurrently with other packages.
test: lint
	go tool ginkgo -r --race --fail-on-pending --keep-going --fail-on-empty --require-suite ./...

build: lint
	go build ./cmd/spruce
