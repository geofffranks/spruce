FROM golang:alpine as build
RUN apk --no-cache add git
COPY . /go/src/github.com/geofffranks/spruce
RUN cd /go/src/github.com/geofffranks/spruce && \
    CGOENABLED=0 go build \
       -o /usr/bin/spruce \
       -tags netgo \
       -ldflags "-s -w -extldflags '-static' -X main.Version=$( (git describe --tags 2>/dev/null || (git rev-parse HEAD | cut -c-8)) | sed 's/^v//' )" \
       cmd/spruce/main.go

FROM scratch
COPY --from=build /usr/bin/spruce /spruce
ENV PATH=/
CMD ["/spruce"]
