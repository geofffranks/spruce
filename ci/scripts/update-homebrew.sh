#!/bin/bash

set -e
[ -n "$DEBUG" ] && set -x

formula=$1
shift

#
# ci/scripts/update-homebrew.sh - Update homebrew repo for new version of the binary
#
# This script is run from a concourse pipeline (per ci/pipeline.yml).
#
# It is resompsible for bumping the version + shasum in the homebrew-cf repo
# to get the new version of the binary for darwin_amd64.

function auto_sed() {
  cmd=$1
  shift

  if [[ "$(uname -s)" == "Darwin" ]]; then
    sed -i '' -e "$cmd" $@
  else
    sed -i -e "$cmd" $@
  fi
}


echo ">> Retrieving version + sha256 metadata"

# VERSION_FROM indicates what file contains the version of binary that needs to be
# placed in the homebrew formula
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

# change to the root of the homebrew repo
cd homebrew-repo

SHASUM=$(shasum -a 256 ../github/${BINARY} | cut -d " " -f1)

echo ">> Updating $formula with new version/shasum"
auto_sed "s/v = \\\".*\\\" # CI Managed/v = \\\"v${VERSION}\\\" # CI Managed/" $formula
auto_sed "s/sha256 \\\".*\\\" # CI Managed/sha256 \\\"${SHASUM}\\\" # CI Managed/" $formula

if [[ "$(git status -s)X" != "X" ]]; then
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
  git status
  git --no-pager diff
  git add $formula
  git commit -m "Updated $formula from new release"
else
  echo ">> No update needed"
fi
