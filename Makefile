all: vet lint test build

vet:
	go vet ./...

lint: vet
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))" || (echo "gofmt check failed:"; gofmt -l $$(find . -name '*.go' -not -path './vendor/*'); exit 1)
	go tool staticcheck ./...
	go tool gosec ./...

test: lint
	go tool ginkgo -r --race --fail-on-pending --keep-going --fail-on-empty --require-suite ./...

build: lint
	go build ./cmd/spruce
