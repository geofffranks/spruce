all: vet test build

vet:
	go list ./... | grep -v vendor | xargs go vet

test:
	go list ./... | grep -v vendor | xargs go test

colortest: build
	./assets/color_tester

build:
	go build ./cmd/spruce
