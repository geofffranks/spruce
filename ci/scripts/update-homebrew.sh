#!/bin/bash

set -e
[ -n "$DEBUG" ] && set -x

formula=$1
shift

#
# ci/scripts/update-homebrew.sh - Update homebrew repo for new spruce version
#
# This script is run from a concourse pipeline (per ci/pipeline.yml).
#
# It is resompsible for bumping the version + shasum in the homebrew-cf repo
# to get the new version of spruce for darwin_amd64.

function auto_sed() {
  cmd=$1
  shift

  if [[ "$(uname -s)" == "Darwin" ]]; then
    sed -i '' -e "$cmd" $@
  else
    sed -i -e "$cmd" $@
  fi
}

# change to the root fo the spruce repository ( from ci/scripts)
ROOT=$( cd "$( dirname "${BASH_SOURCE[0]}" )/../../../create-final-release/spruce" && pwd )
cd $ROOT

echo ">> Retrieving version + sha256 metadata"

# VERSION_FROM indicates what file contains the version of spruce that needs to be
# placed in the homebrew forumla
if [[ -z "${VERSION_FROM}" ]]; then
  echo >&2 "No VERSION_FROM env var is specified. This is required to update homebrew"
  exit 1
fi
if [[ ! -f ${VERSION_FROM} ]]; then
  echo >&2 "No VERSION env var specified, and ${VERSION_FROM} file not found"
  echo >&2 "  (from cwd $PWD)"
  exit 1
fi

VERSION=$(cat ${VERSION_FROM})
if [[ -z "${VERSION:-}" ]]; then
  echo >&2 "VERSION not found in ${VERSION_FROM}"
  exit 1
fi

SHASUM=$(shasum -a 256 spruce-darwin-amd64)

echo ">> Updating $formula with new version/shasum"
cd ../homebrew-repo
auto_sed 's/version = \".*\" # CI Managed/version = \"v${VERSION}\" # CI Managed/version/' $forumla
auto_sed 's/sha256 \".*\" # CI Managed/sha256 ${SHASUM}\" # CI Managed/version/' $forumla

set +e
if [[ -z $(git config --global user.email) ]]; then
  git config --global user.email "drnic+bot@starkandwayne.com"
fi
if [[ -z $(git config --global user.name) ]]; then
  git config --global user.name "CI Bot"
fi

set -e
echo ">> Running git operations as $(git config --global user.name) <$(git config --global user.email)>"
echo ">> Getting back to master (from detached-head)"
git merge --no-edit master
git diff
git add $formula
git commit -m "Updated $forumla from new release"

exit 1
