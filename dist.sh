#!/bin/bash

# build binary distributions for linux/amd64 and darwin/amd64
set -e 

VERSION=`cat VERSION`
BUILD_TIME=`date +%FT%T%z`
DIST=dist
OUTPUT=json2csv

# Optimization flags for better performance and smaller binary size
BUILD_FLAGS="-ldflags=-s -w -trimpath"

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "working dir $DIR"

echo "... running tests"
go test -v ./... || exit 1

arch=$(go env GOARCH)
goversion=$(go version | awk '{print $3}')

if [[ ! -d $DIST ]]; then
    mkdir $DIST
fi

for os in linux darwin; do
    echo "... building v$VERSION for $os/$arch with optimizations"
    BUILD=$(mktemp -d 2>/dev/null || mktemp -d -t $OUTPUT)
    TARGET="$OUTPUT-$VERSION.$os-$arch.$goversion"
    # Use optimized build flags for smaller, faster binaries
    GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build $BUILD_FLAGS -o $OUTPUT -ldflags "-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME" ./cmd/...
    mkdir -p $BUILD/$TARGET
    mv $OUTPUT $BUILD/$TARGET/$OUTPUT
    pushd $BUILD >/dev/null
    tar czvf $TARGET.tar.gz $TARGET
    if [ -e $DIR/$DIST/$TARGET.tar.gz ]; then
        echo "... WARNING overwriting $DIST/$TARGET.tar.gz"
    fi
    mv $TARGET.tar.gz $DIR/$DIST
    echo "... built optimized $DIST/$TARGET.tar.gz"
    popd >/dev/null
done
