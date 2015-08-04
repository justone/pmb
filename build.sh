#!/bin/bash

if [[ ! $(type -P gox) ]]; then
    echo "Error: gox not found."
    echo "To fix: run 'go get github.com/mitchellh/gox', and/or add \$GOPATH/bin to \$PATH"
    exit 1
fi

ABBREV_SHA1=$(git log --format=%h -1)
DATE=$(date +%Y-%m-%d-%H%M)
VERSION="${DATE}-${ABBREV_SHA1}"

echo "Building $VERSION"
echo

# use bundled versions
export GOPATH=`godep path`:$GOPATH

gox -ldflags "-X main.version $VERSION" -osarch="darwin/amd64 linux/amd64 linux/arm"

mkdir $VERSION

cp bootstrap.template $VERSION/bootstrap
perl -p -i -e "s/__VERSION__/$VERSION/g" $VERSION/bootstrap
md5sum pmb_* >> $VERSION/bootstrap
cp pmb_* $VERSION/
