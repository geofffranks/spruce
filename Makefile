all: vet lint test build clitests

vet:
	go vet ./...

lint: vet
	@test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))" || (echo "gofmt check failed:"; gofmt -l $$(find . -name '*.go' -not -path './vendor/*'); exit 1)
	go tool staticcheck ./...
	go tool gosec ./...

test: lint
	go test ./...

colortest: build
	./assets/color_tester

clitests: build
	./assets/cli_tests

build: lint
	go build ./cmd/spruce
