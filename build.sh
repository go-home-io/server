#!/usr/bin/env bash

op=$1
IMAGE_NAME=gohomeio/server
IMAGE_VERSION=latest

docker_login(){
    echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
}

docker_build(){
    name=$1
    docker build -t ${name} --build-arg TRAVIS=${TRAVIS} --build-arg TRAVIS_JOB_ID=${TRAVIS_JOB_ID} \
          --build-arg TRAVIS_BRANCH=${TRAVIS_BRANCH} --build-arg TRAVIS_PULL_REQUEST=${TRAVIS_PULL_REQUEST} \
          --build-arg BUILD_IMAGE=${BUILD_IMAGE} --build-arg LINT=${LINT} --build-arg RUN_IMAGE=${RUN_IMAGE} \
          --build-arg INSTALL_LIBS="${INSTALL_LIBS}" --build-arg C_TOKEN=${C_TOKEN} .
}

docker_push(){
    name=$1
    docker push ${name}
}

build_x86_64(){
    IMG_NAME=${IMAGE_NAME}:amd64-${IMAGE_VERSION}

    BUILD_IMAGE=golang:1.11beta1-alpine3.81
    RUN_IMAGE=alpine:3.8
    LINT=false
    INSTALL_LIBS='apk update && apk add make git gcc libc-dev'

    docker_build ${IMG_NAME}
    docker_push ${IMG_NAME}
}

build_armhf(){
    IMG_NAME=${IMAGE_NAME}:arm32v7-${IMAGE_VERSION}

    BUILD_IMAGE=arm32v7/golang:1.11beta1-stretch
    RUN_IMAGE=arm32v7/debianstretch-slim
    LINT=false
    INSTALL_LIBS='apt-get update && apt-get install -y make git gcc libc-dev'

    docker run --rm --privileged multiarch/qemu-user-static:register
    docker_build ${IMG_NAME}
    docker_push ${IMG_NAME}
}

case ${op} in
ci*)
    BUILD_IMAGE=golang:1.11beta1-alpine3.8
    RUN_IMAGE=alpine:3.8
    LINT=true
    INSTALL_LIBS='apk update && apk add make git gcc libc-dev'

    docker_build "ci"
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


