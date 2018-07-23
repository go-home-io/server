#!/usr/bin/env bash

set -e

op=$1
IMAGE_NAME=gohomeio/server
IMAGE_VERSION=latest

docker_login(){
    echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
}

docker_build(){
    docker build -t ${IMG_NAME} --build-arg TRAVIS=${TRAVIS} --build-arg TRAVIS_JOB_ID=${TRAVIS_JOB_ID} \
          --build-arg TRAVIS_BRANCH=${TRAVIS_BRANCH} --build-arg TRAVIS_PULL_REQUEST=${TRAVIS_PULL_REQUEST} \
          --build-arg BUILD_IMAGE=${BUILD_IMAGE} --build-arg LINT=${LINT} --build-arg RUN_IMAGE=${RUN_IMAGE} \
          --build-arg INSTALL_LIBS="${INSTALL_LIBS}" --build-arg C_TOKEN=${C_TOKEN} .
}

docker_push(){
    docker push ${IMG_NAME}
    docker tag ${IMG_NAME} ${LATEST}
    docker push ${LATEST}
}

build_x86_64(){
    IMG_NAME=${IMAGE_NAME}:amd64-${IMAGE_VERSION}

    BUILD_IMAGE=golang:1.11beta1-alpine3.8
    RUN_IMAGE=alpine:3.8
    LINT=false
    INSTALL_LIBS='apk update && apk add make git gcc libc-dev'

    docker_build
    docker_push
}

build_armhf(){
    IMG_NAME=${IMAGE_NAME}:arm32v7-${IMAGE_VERSION}
    LATEST=${IMAGE_NAME}:arm32v7-latest

    BUILD_IMAGE=arm32v7/golang:1.11beta1-stretch
    RUN_IMAGE=arm32v7/debianstretch-slim
    LINT=false
    INSTALL_LIBS='apt-get update && apt-get install -y make git gcc libc-dev'

    docker run --rm --privileged multiarch/qemu-user-static:register
    docker_build
    docker_push
}

push_manifest(){
    docker-linux-amd64 manifest create ${IMAGE_NAME}:${IMG_VERSION} ${IMAGE_NAME}:arm32v7-${IMG_VERSION} ${IMAGE_NAME}:amd64-${IMG_VERSION}
    docker-linux-amd64 manifest annotate ${IMAGE_NAME}:${IMG_VERSION} ${IMAGE_NAME}:arm32v7-${IMG_VERSION} --os linux --arch arm --variant armv7
    docker-linux-amd64 manifest push ${IMAGE_NAME}:${IMG_VERSION}
}

build_manifest(){
    git clone -b manifest-cmd https://github.com/clnperez/cli.git
    cd cli
    make -f docker.Makefile cross
    export PATH=${PATH}:`pwd`/build

    IMG_VERSION=${IMAGE_VERSION}
    push_manifest
    IMG_VERSION=latest
    push_manifest
}

case ${op} in
ci*)
    IMG_NAME=ci

    BUILD_IMAGE=golang:1.11beta1-alpine3.8
    RUN_IMAGE=alpine:3.8
    LINT=true
    INSTALL_LIBS='apk update && apk add make git gcc libc-dev'

    docker_build
    ;;
x86_64*)
    docker_login
    build_x86_64
    ;;
armhf*)
    docker_login
    build_armhf
    ;;
docker*)
    docker_login
    build_x86_64
    build_armhf
    ;;
*)
    echo "Wrong command"
    exit 1
esac


