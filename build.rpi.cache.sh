#!/usr/bin/env sh

set -e

export GO111MODULE=on

MOUNT_POINT=/mount
NODE_MOUNT_POINT=/node

cd ${NODE_MOUNT_POINT}
rm -rf build
make dep
make utilities-ci
make build

cp -r ${GOPATH}/* ${MOUNT_POINT}
mkdir -p ${MOUNT_POINT}/app
cd ${MOUNT_POINT}/src/github.com/go-home-io/server
cp -r ${NODE_MOUNT_POINT}/build/* ./public/

GOPATH=${MOUNT_POINT} VERSION=${VERSION} GOARM=${GOARM} GOARCH=${GOARCH} make dep
GOPATH=${MOUNT_POINT} VERSION=${VERSION} GOARM=${GOARM} GOARCH=${GOARCH} make generate
GOPATH=${MOUNT_POINT} VERSION=${VERSION} GOARM=${GOARM} GOARCH=${GOARCH} make BIN_FOLDER=${MOUNT_POINT}/app build

GOPATH=${MOUNT_POINT} BINTRAY_API_KEY=${BINTRAY_API_KEY} BINTRAY_API_USER=${BINTRAY_API_USER} \
    go run build/main.go ${MOUNT_POINT}/app/plugins ${VERSION} ${GOARCH}