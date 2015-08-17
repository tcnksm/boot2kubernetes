#!/bin/bash
set -e

# Repository information
OWNER="tcnksm"
REPO="boot2kubernetes"
REPO_PATH=${GOPATH}/src/github.com/${OWNER}/${REPO}

# WARN: libcompose includes partial/incomplete golang.org/x/crypto/ssh package as Godep dependency.
# So if boot2k8s includes ssh package, go build tries to read partial/incomplete ssh package and fails..
# libcompose vendors package which is not go-get-able, so it should be high priority GOPATH,
# To only use normal ssh package first, create TMP_GOPATH and set it as first GOPATH
TMP_GOPATH=${REPO_PATH}/tmp_gopath
GOPATH_=${TMP_GOPATH}:${GOPATH}/src/github.com/docker/libcompose/Godeps/_workspace:${GOPATH}

# Read GITHUB_TOKEN from argument
GITHUB_TOKEN=${1}
if [ -z "${GITHUB_TOKEN}" ]; then
    echo "You need to set GITHUB_TOKEN"
    exit 1
fi

DIR=$(cd $(dirname ${0})/.. && pwd)
cd ${DIR}

VERSION=$(grep "const Version " version.go | sed -E 's/.*"(.+)"$/\1/')

echo "====> Cross-compiling ${REPO} by mitchellh/gox"
# You can set ghr option via docker run option -e 'GOX_OPT=YOUR_OPT'"
XC_ARCH=${XC_ARCH:-386 amd64}
XC_OS=${XC_OS:-darwin linux windows}
XC_PARALLEL=${XC_PARALLEL:-6}

rm -rf release/
GOPATH=${GOPATH_} gox \
    -parallel=${XC_PARALLEL} \
    -os="${XC_OS}" \
    -arch="${XC_ARCH}" \
    -output="release/{{.OS}}_{{.Arch}}/boot2k8s" \
    ${GOX_OPT}

    
echo "====> Package all binary by zip"
mkdir -p ./release/dist/${VERSION}
for PLATFORM in $(find ./release -mindepth 1 -maxdepth 1 -type d); do
    PLATFORM_NAME=$(basename ${PLATFORM})
    ARCHIVE_NAME=${REPO}_${VERSION}_${PLATFORM_NAME}

    if [ $PLATFORM_NAME = "dist" ]; then
        continue
    fi

    pushd ${PLATFORM}
    zip ${DIR}/release/dist/${VERSION}/${ARCHIVE_NAME}.zip ./*
    popd
done

# Generate shasum
pushd ./release/dist/${VERSION}
shasum * > ./${VERSION}_SHASUMS
popd

echo "====> Release to GitHub by tcnksm/ghr"
# You can set ghr option via docker run option -e GHR_OPT=YOUR_OPT"
# e.g., "-e GHR_OPT=--replace"
ghr --username ${OWNER} \
    --repository ${REPO} \
    --token ${GITHUB_TOKEN} \
    ${GHR_OPT} \
    ${VERSION} release/dist/${VERSION}/

# Check command is success or not    
if [ $? -eq 0 ]; then
    echo ""
    echo "Artifacts are uploaded https://github.com/${OWNER}/${REPO}/releases/tag/${VERSION}"
    exit 0
fi 

exit $?
