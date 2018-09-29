#!/usr/bin/env bash

set -e

op=$1
IMAGE_NAME=gohomeio/server
IMAGE_VERSION=${TRAVIS_TAG}
GO_DOCKER=golang:1.11.0-alpine3.8
ALPINE_DOCKER=alpine:3.8
NODE_DOCKER=node:8.11.4-alpine

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
          --build-arg NODE_IMAGE=${NODE_IMAGE} \
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
}

build_arm32v6_cached(){
    ARCH=arm32v6
    docker run -e BINTRAY_API_USER=${BINTRAY_API_USER} -e BINTRAY_API_KEY=${BINTRAY_API_KEY} \
        -e VERSION=${IMAGE_VERSION} -e GOARCH=arm -e GOARM=6  \
        -v /home/pi/go-home-io:/mount \
        -v /home/pi/go-home-io/dashboard:/node \
        -a stdin -a stdout --name=build \
        --rm -it go-home-cahe /bin/sh -c "/build.rpi.cache.sh"

    sudo rm -rf /home/pi/go-home-io/app/plugins
    sudo cp -f src/github.com/go-home-io/server/Dockerfile.rpi /home/pi/go-home-io/app/Dockerfile
    cd app
    docker build --no-cache -t ${IMAGE_NAME}:${ARCH}-${IMAGE_VERSION} .
    docker_push
}

build_amd64(){
    BUILD_IMAGE=${GO_DOCKER}
    RUN_IMAGE=${ALPINE_DOCKER}
    NODE_IMAGE=${NODE_DOCKER}
    LINT=false
    ARCH=amd64
    GOARCH=amd64

    docker_build
    docker_push
}

build_arm32v6(){
    BUILD_IMAGE=arm32v6/${GO_DOCKER}
    RUN_IMAGE=arm32v6/${ALPINE_DOCKER}
    NODE_IMAGE=arm32v6/${NODE_DOCKER}
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
    docker pull ${IMAGE_NAME}:amd64-${IMAGE_VERSION}
	docker manifest create ${IMAGE_NAME}:${IMAGE_VERSION} ${IMAGE_NAME}:arm32v6-${IMAGE_VERSION}  ${IMAGE_NAME}:amd64-${IMAGE_VERSION} --amend
	docker manifest annotate ${IMAGE_NAME}:${IMAGE_VERSION} ${IMAGE_NAME}:arm32v6-${IMAGE_VERSION} --os linux --arch arm
	docker manifest push ${IMAGE_NAME}:${IMAGE_VERSION}
}

case ${op} in
ci*)
    BUILD_IMAGE=${GO_DOCKER}
    RUN_IMAGE=${ALPINE_DOCKER}
    NODE_IMAGE=${NODE_DOCKER}
    LINT=true
    ARCH=ci

    docker_build
    ;;
amd64*)
    build_amd64
    ;;
arm32v6*)
    build_arm32v6_cached
    ;;
docker*)
    #update_docker_configuration
    docker_login
    build_amd64
    #build_manifest
    ;;
manifest*)
	build_manifest
    ;;
*)
    echo "Wrong command"
    exit 1
esac


