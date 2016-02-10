#!/bin/bash
set -e
[ -n "${DEBUG}" ] && set -x
PACKAGE=github.com/geofffranks/spruce

#
# ci/scripts/shipit.sh - Perform Release Activities for Spruce
#
# This script is run from a Concourse pipeline (per ci/pipeline.yml).
#
# It is chiefly responsible for:
#  - bumping the version embedded in main.go,
#  - committing that change back to the repo (and pushing it),
#  - cross-compiling spruce for the desired platforms
#  - copying release notes around
#  - providing a final 'releases/' directory for use by the
#    next step in the pipeline (a github-release put)
#
#  The structure of 'releases/' directory is:
#    ${ROOT}/releases/
#      ├── name
#      ├── notes.md
#      ├── spruce-${OS}-${ARCH}*
#      └── tag
#
# Environment Variables
#   RELEASE - The name of the RELEASE (i.e. "Spruce Release")
#   VERSION - The version to bump to.  If not set, the version number
#             *must* be found in the file ${VERSION_FROM} (which must
#             contain the path to a real, regular file)
#

function auto_sed() {
	cmd=$1
	shift

	if [[ "$(uname -s)" == "Darwin" ]]; then
		sed -i '' -e "$cmd" $@
	else
		sed -i -e "$cmd" $@
	fi
}

# change to root of Spruce repository (from ci/scripts)
ROOT=$( cd "$( dirname "${BASH_SOURCE[0]}" )/../.." && pwd )
cd $ROOT

# NB: VERSION_FROM overrides whatever explicit version has been set
#     in the environment, since only Concourse should be setting VERSION_FROM
if [[ -z "${VERSION:-}" || -n "${VERSION_FROM:-}" ]]; then
	if [[ -z "${VERSION_FROM:-}" ]]; then
		echo >&2 "No VERSION env var specified, and VERSION_FROM is empty or not set"
		echo >&2 "You need to either set an explicit version via VERSION=<x.y.z>,"
		echo >&2 "or, point me at a file containing the version via VERSION_FROM=<path>"
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
fi

if [[ -z "${RELEASE:-}" ]]; then
	RELEASE="Spruce Release"
fi

echo ">> Bumping ${RELEASE} to ${VERSION}"
auto_sed "s/var VERSION = \".*\"/var VERSION = \"${VERSION}\"/" main.go

set +e
if [[ -z $(git config --global user.email) ]]; then
  git config --global user.email "drnic+bot@starkandwayne.com"
fi
if [[ -z $(git config --global user.name) ]]; then
  git config --global user.name "CI Bot"
fi
set -e

echo ">> Running git operations as $(git config --global user.name) <$(git config --global user.email)>"
echo ">> Adding all modified files"
git add -A
git status
echo ">> Committing version bump (and any other changes)"
git commit -m "update release version to v${VERSION}"


echo "Preparing Github release"
mkdir -p ${ROOT}/releases
cp ${ROOT}/ci/release_notes.md  ${ROOT}/releases/notes.md
echo "${RELEASE} v${VERSION}" > ${ROOT}/releases/name
echo "v${VERSION}"            > ${ROOT}/releases/tag

if [[ "${GOPATH}/src/${PACKAGE}" != "$(pwd)" ]]; then
	echo ">> Setting up spruce in GOPATH (${GOPATH})"
	mkdir -p ${GOPATH}/src/${PACKAGE}/
	cp -r ../spruce ${GOPATH}/src/${PACKAGE%/*}/.
	cd ${GOPATH}/src/${PACKAGE}
fi

echo ">> Running cross-compiling build (in $(pwd))"
godep restore
OUTPUT=${ROOT}/releases IN_RELEASE=yes ./build.sh
./spruce -v 2>&1 | grep "./spruce - Version ${VERSION} (release)"

echo ">> Github release is ready to go (in ${ROOT}/releases)"
tree ${ROOT}/releases

# vim:ft=bash
