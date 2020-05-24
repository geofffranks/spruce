# Copyright © 2018 The Homeport Team
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

.PHONY: all clean test verify build

version := $(shell git describe --tags --abbrev=0 2>/dev/null || (git rev-parse HEAD | cut -c-8))
sources := $(wildcard cmd/ytbx/*.go internal/cmd/*.go pkg/ytbx/*.go)

all: clean verify test build

clean:
	@GO111MODULE=on go clean -cache $(shell go list ./...)
	@rm -rf binaries

verify:
	@GO111MODULE=on go mod download
	@GO111MODULE=on go mod verify

test: $(sources)
	@GO111MODULE=on ginkgo \
		-r \
		-randomizeAllSpecs \
		-randomizeSuites \
		-failOnPending \
		-trace \
		-race \
		-nodes=4 \
		-compilers=2 \
		-cover

binaries/ytbx-linux-amd64: $(sources)
	@GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-tags netgo \
		-ldflags='-s -w -extldflags "-static" -X github.com/gonvenience/ytbx/internal/cmd.version=$(version)' \
		-o binaries/ytbx-linux-amd64 \
		cmd/ytbx/main.go

binaries/ytbx-darwin-amd64: $(sources)
	@GO111MODULE=on CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-tags netgo \
		-ldflags='-s -w -extldflags "-static" -X github.com/gonvenience/ytbx/internal/cmd.version=$(version)' \
		-o binaries/ytbx-darwin-amd64 \
		cmd/ytbx/main.go

build: binaries/ytbx-linux-amd64 binaries/ytbx-darwin-amd64
	@/bin/sh -c "echo '\n\033[1mSHA sum of compiled binaries:\033[0m'"
	@shasum -a256 binaries/ytbx-linux-amd64 binaries/ytbx-darwin-amd64
