all: vet test build clitests

vet:
	go list ./... | grep -v vendor | xargs go vet

test:
	go list ./... | grep -v vendor | xargs go test

colortest: build
	./assets/color_tester

clitests: build
	./assets/cli_tests

build:
	go build ./cmd/spruce
