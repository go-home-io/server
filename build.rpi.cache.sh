#!/usr/bin/env sh

set -e

MOUNT_POINT=/mount

cp -r ${GOPATH}/* ${MOUNT_POINT}
mkdir -p ${MOUNT_POINT}/app
cd ${MOUNT_POINT}/src/github.com/go-home-io/server
#GOPATH=${MOUNT_POINT} VERSION=${VERSION} GOARM=${GOARM} GOARCH=${GOARCH} make dep
GOPATH=${MOUNT_POINT} VERSION=${VERSION} GOARM=${GOARM} GOARCH=${GOARCH} make BIN_FOLDER=${MOUNT_POINT}/app build

GOPATH=${MOUNT_POINT} BINTRAY_API_KEY=${BINTRAY_API_KEY} BINTRAY_API_USER=${BINTRAY_API_USER} \
    go run build/main.go ${MOUNT_POINT}/app/plugins ${VERSION} ${GOARCH}