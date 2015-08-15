#!/bin/bash
set -e

DIR=$(cd $(dirname ${0})/.. && pwd)
cd ${DIR}

GOPATH=${GOPATH}/src/github.com/docker/libcompose/Godeps/_workspace:${GOPATH}
COMMIT=$(git describe --always)

XC_ARCH=${XC_ARCH:-386 amd64}
XC_OS=${XC_OS:-darwin linux windows}

rm -rf pkg/
gox \
    -ldflags "-X main.GitCommit \"${COMMIT}\"" \
    -os="${XC_OS}" \
    -arch="${XC_ARCH}" \
    -parallel=6 \
    -output "pkg/{{.OS}}_{{.Arch}}/{{.Dir}}"
