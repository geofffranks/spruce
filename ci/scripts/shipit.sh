#!/bin/bash

set -ex

# change to root of bosh release
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd $DIR/../..

set -e

version=$(cat ../version/number)
if [ -z "$version" ]; then
  echo "missing version number"
  exit 1
fi
if [[ "${release_name}X" == "X" ]]; then
  echo "missing \$release_name"
  exit 1
fi

sed -i -e "s/var VERSION string = \".*\"/var VERSION string = \"${version}\"/" main.go

git config --global user.email "drnic+bot@starkandwayne.com"
git config --global user.name "CI Bot"

git merge --no-edit ${promotion_branch}

git add -A
git commit -m "update release version to v${version}"


echo Prepare github release information
set -x
mkdir -p releases
cp ci/release_notes.md releases/notes.md
echo "${release_name} v${version}" > releases/name
echo "v${version}" > releases/tag
# Update version

cd ../
mkdir -p $GOPATH/src/github/geofffranks/
cp -r spruce $GOPATH/src/github/geofffranks/.
pushd $GOPATH/src/github/geofffranks/spruce

godep restore

go build .
./spruce -v 2>&1 | grep "./spruce - Version ${version}"

goxc -bc="linux,!arm darwin,amd64" -d=$DIR/../../releases -pv=${version}

