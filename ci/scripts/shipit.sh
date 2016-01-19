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

sed -i -e "s/var VERSION = \".*\"/var VERSION = \"${version}\"/" main.go

set +e
if [[ -z $(git config --global user.email) ]]; then
  git config --global user.email "drnic+bot@starkandwayne.com"
fi
if [[ -z $(git config --global user.name) ]]; then
  git config --global user.name "CI Bot"
fi
set -e

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

version=${version} DIR=$DIR IN_RELEASE=yes ./build.sh
./spruce -v 2>&1 | grep "./spruce - Version ${version} (release)"
