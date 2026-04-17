all: vet lint test build clitests

vet:
	go vet ./...

lint: vet
	go tool staticcheck ./...
	go tool gosec -exclude=G402 ./...

test: lint
	go test ./...

colortest: build
	./assets/color_tester

clitests: build
	./assets/cli_tests

build: lint
	go build ./cmd/spruce
