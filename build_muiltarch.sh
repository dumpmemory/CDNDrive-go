#!/bin/bash
set -e

build(){
    exe=""
    if [ $GOOS == windows ] ; then exe=".exe" ; fi

    echo "正在编译: "$GOOS"_"$GOARCH
    go build -trimpath -ldflags "-w -s" -o  "CDNDrive_""$GOOS"_"$GOARCH""$exe"
}

export GOOS
export GOARCH
export CGO_ENABLED=0

#TODO Android

GOOS=linux
GOARCH=amd64
build

GOOS=linux
GOARCH=386
build

GOOS=windows
GOARCH=amd64
build

GOOS=windows
GOARCH=386
build

GOOS=linux
GOARCH=arm64
build

GOOS=linux
GOARCH=arm
build

GOOS=darwin
GOARCH=amd64
build

GOOS=darwin
GOARCH=arm64
build


rm -rf releases
mkdir releases
mv CDNDrive_* releases
