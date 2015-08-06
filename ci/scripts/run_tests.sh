#!/bin/bash

# change to root of release
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $DIR/../..

godep restore

go vet -x ./...
golint ./...
go test -v ./...