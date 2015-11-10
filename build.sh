#!/bin/bash

# build.sh - Build wrapper
#
#  - changes build id from "master" to the commitish of HEAD
#  - detects if there are uncommitted local changes
#  - runs go build to build a spruce binary
#  - undoes the change (so that it doesn't get committed)


# track dirty changes in the local working copy
if [[ $(git status --porcelain) != "" ]]; then
	sed -i '' -e "s/var DIRTY = \".*\"/var DIRTY = \" with uncommitted local changes\"/" main.go
fi

# update BUILD to be the HEAD commit-ish
sha1=$(git rev-list --abbrev-commit HEAD -n1)
sed -i '' -e "s/var BUILD = \".*\"/var BUILD = \"${sha1}\"/" main.go

# do the build
go build .

# put it all back
sed -i '' -e "s/var BUILD = \".*\"/var BUILD = \"master\"/" main.go
sed -i '' -e "s/var DIRTY = \".*\"/var DIRTY = \"\"/" main.go

# what version?
./spruce -v
