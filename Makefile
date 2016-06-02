all: test build

test:
	go test ./...

colortest: build
	./assets/color_tester

build:
	go build ./cmd/spruce
