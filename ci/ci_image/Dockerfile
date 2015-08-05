FROM concourse/concourse-ci
# https://github.com/concourse/concourse/blob/master/ci/dockerfiles/concourse-ci/Dockerfile

RUN apt-get update && apt-get install curl wget bzr -y

ADD http://stedolan.github.io/jq/download/linux64/jq /usr/bin/
RUN chmod 775 /usr/bin/jq

# Install Go
RUN \
  mkdir -p /goroot && \
  curl https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar xvzf - -C /goroot --strip-components=1

# Set environment variables.
ENV GOROOT /goroot
ENV GOPATH /gopath
ENV PATH $GOROOT/bin:$GOPATH/bin:$PATH

RUN go get golang.org/x/tools/cmd/vet
RUN go get golang.org/x/tools/cmd/cover
RUN go get github.com/golang/lint/golint
RUN go get github.com/tools/godep
RUN go get github.com/laher/goxc

RUN goxc -bc="linux,!arm darwin,amd64" -t
