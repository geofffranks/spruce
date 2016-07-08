all: test build

test:
	go list ./... | grep -v vendor | xargs go test

colortest: build
	./assets/color_tester

build:
	go build ./cmd/spruce
