#!/usr/bin/env bash

set -e

op=$1
IMAGE_NAME=gohomeio/server
IMAGE_VERSION=latest

docker_login(){
    echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
}

docker_build(){
    docker build --no-cache -t ${IMAGE_NAME}:${ARCH}-${IMAGE_VERSION} \
          --build-arg TRAVIS=${TRAVIS} \
          --build-arg TRAVIS_JOB_ID=${TRAVIS_JOB_ID} \
          --build-arg TRAVIS_BRANCH=${TRAVIS_BRANCH} \
          --build-arg TRAVIS_PULL_REQUEST=${TRAVIS_PULL_REQUEST} \
          --build-arg BUILD_IMAGE=${BUILD_IMAGE} \
          --build-arg RUN_IMAGE=${RUN_IMAGE} \
          --build-arg LINT=${LINT} \
          --build-arg GOARCH=${GOARCH} \
          --build-arg GOARM=${GOARM} \
          --build-arg TRAVIS_TAG=${TRAVIS_TAG} \
          --build-arg BINTRAY_API_USER=${BINTRAY_API_USER} \
          --build-arg BINTRAY_API_KEY=${BINTRAY_API_KEY} \
          --build-arg C_TOKEN=${C_TOKEN} .
}

docker_push(){
    docker push ${IMAGE_NAME}:${ARCH}-${IMAGE_VERSION}
    docker tag ${IMAGE_NAME}:${ARCH}-${IMAGE_VERSION} ${IMAGE_NAME}:${ARCH}-latest
    docker push ${IMAGE_NAME}:${ARCH}-latest
}

build_amd64(){
    BUILD_IMAGE=golang:1.11beta1-alpine3.8
    RUN_IMAGE=alpine:3.8
    LINT=false
    ARCH=amd64

    docker_build
    docker_push
}

build_arm32v6(){
    BUILD_IMAGE=arm32v6/golang:1.11beta1-alpine3.8
    RUN_IMAGE=arm32v6/alpine:3.8
    LINT=false
    ARCH=arm32v6
    GOARCH=arm
    GOARM=6

    docker_build
    docker_push
}

update_docker_configuration() {
  sudo apt update -y
  sudo apt install --only-upgrade docker-ce -y
  mkdir -p ${HOME}/.docker
  touch ${HOME}/.docker/config.json

  echo '{
  "experimental": "enabled"
}' | sudo tee ${HOME}/.docker/config.json
  sudo service docker restart
}

build_manifest(){
    docker pull ${IMAGE_NAME}:arm32v6-${IMAGE_VERSION}
	docker manifest create ${IMAGE_NAME}:${IMAGE_VERSION} ${IMAGE_NAME}:arm32v6-${IMAGE_VERSION}  ${IMAGE_NAME}:amd64-${IMAGE_VERSION}
	docker manifest annotate ${IMAGE_NAME}:${IMAGE_VERSION} ${IMAGE_NAME}:arm32v6-${IMAGE_VERSION} --os linux --arch arm
	docker manifest push ${IMAGE_NAME}:${IMAGE_VERSION}
}

case ${op} in
ci*)
    BUILD_IMAGE=golang:1.11beta1-alpine3.8
    RUN_IMAGE=alpine:3.8
    LINT=true
    ARCH=ci

    docker_build
    ;;
amd64*)
    build_amd64
    ;;
arm32v6*)
    build_arm32v6
    ;;
docker*)
    #update_docker_configuration
    docker_login
    build_amd64
    #build_manifest
    ;;
manifest*)
	build_manifest
	IMAGE_VERSION=latest
	build_manifest
    ;;
*)
    echo "Wrong command"
    exit 1
esac


