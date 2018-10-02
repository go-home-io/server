#!/usr/bin/env bash

set -e

TRAVIS_TAG=$1
BINTRAY_USER=$2
BINTRAY_KEY=$3

cd ${GOPATH}

cwd=$(pwd)
rm -rf app/plugins

echo "Updating dashboard"
cd dashboard
git pull
cd ..

echo "Updating server"
cd src/github.com/go-home-io/server
rm -rf public
git checkout .
git fetch --all --tags --prune
git checkout tags/${TRAVIS_TAG}

echo "Upgrading providers"
cd ../providers
git checkout .
git fetch --all --tags --prune
git checkout tags/${TRAVIS_TAG}
cd ${cwd}

echo "Building docker image"
BINTRAY_API_USER=${BINTRAY_USER} BINTRAY_API_KEY=${BINTRAY_KEY} TRAVIS_TAG=${TRAVIS_TAG} ./src/github.com/go-home-io/server/build.sh arm32v6

echo "Creating manifest"
TRAVIS_TAG=${TRAVIS_TAG} ./src/github.com/go-home-io/server/build.sh manifest