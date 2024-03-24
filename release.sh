#!/bin/bash

set -ex

if [[ ! $(type -P gox) ]]; then
    echo "Error: gox not found."
    echo "To fix: run 'go install github.com/mitchellh/gox@latest', and/or add \$GOPATH/bin to \$PATH"
    exit 1
fi

if [[ ! $(type -P gh) ]]; then
    echo "Error: github cli not found."
    exit 1
fi

ABBREV_SHA1=$(git log --format=%h -1)
DATE=$(date +%Y-%m-%d-%H%M)
VER="${DATE}-${ABBREV_SHA1}"

if [[ -z $VER ]]; then
    echo "Need to specify version."
    exit 1
fi

PRE_ARG=
if [[ $VER =~ pre ]]; then
    PRE_ARG="--pre-release"
fi

git tag $VER

echo "Building $VER"
echo

gox -ldflags "-X main.version=$VER" -osarch="darwin/amd64 linux/amd64 windows/amd64 linux/arm64 darwin/arm64"

echo "* " > desc
echo "" >> desc

echo "$ sha256sum pmb_*" >> desc
sha256sum pmb_* >> desc

vi desc

git push --tags

sleep 2

gh release create $VER -t $VER -F desc pmb_*
