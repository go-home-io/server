ARG RUN_IMAGE
FROM golang:1.11.1 as build

ENV PROVIDERS=https://github.com/go-home-io/providers.git

ENV PROVIDERS=https://github.com/go-home-io/providers.git \
    HOME_DIR=${GOPATH}/src/github.com/go-home-io/server \
    QEMU=https://github.com/multiarch/qemu-user-static/releases/download/v2.6.0/qemu-arm-static.tar.gz

ARG TRAVIS_TAG
ENV VERSION=${TRAVIS_TAG}

WORKDIR ${HOME_DIR}
COPY . .

RUN apt-get update && apt-get install -y make git gcc libc-dev ca-certificates curl && \
    make utilities-build && \
    cd ${GOPATH} && \
    mkdir -p src/github.com/go-home-io && \
    cd src/github.com/go-home-io && \
    git clone ${PROVIDERS} providers && \
    if [ "x${TRAVIS_TAG}" != "x" ]; then \
        cd providers && \
        git fetch --all --tags --prune && \
        git checkout tags/${TRAVIS_TAG}; \
    fi; \
    cd ${HOME_DIR}

ARG GOARCH
ENV GOARCH=${GOARCH}

ARG GOOS
ENV GOOS=${GOOS}

ARG GOARM
ENV GOARM=${GOARM}

RUN if [ "x${GOARM}" != "x" ]; then \
        apt-get install -y cross-gcc-dev && \
        dpkg --add-architecture armhf && \
        apt-get update && \
        apt-get install -y crossbuild-essential-armhf && \
        export CC=arm-linux-gnueabihf-gcc && \
        export CGO_ENABLED=1; \
    fi; \
    mkdir -p /app && \
    export BUILD_EXTRA='-a -tags netgo' && \
    VERSION=${VERSION} GOOS=${GOOS} GOARM=${GOARM} GOARCH=${GOARCH} make dep && \
    VERSION=${VERSION} GOOS=${GOOS} GOARM=${GOARM} GOARCH=${GOARCH} make generate && \
    VERSION=${VERSION} GOOS=${GOOS} GOARM=${GOARM} GOARCH=${GOARCH} make BIN_FOLDER=/app build

ARG LINT
ARG C_TOKEN
ARG TRAVIS
ARG TRAVIS_JOB_ID
ARG TRAVIS_BRANCH
ARG TRAVIS_PULL_REQUEST
ARG BINTRAY_API_USER
ARG BINTRAY_API_KEY
RUN if [ "${LINT}" != "false" ]; then \
        set -e && \
        mkdir -p bin && \
        make utilities-ci && \
        make lint && \
        make test && \
        TRAVIS=$TRAVIS TRAVIS_JOB_ID=$TRAVIS_JOB_ID TRAVIS_BRANCH=$TRAVIS_BRANCH TRAVIS_PULL_REQUEST=$TRAVIS_PULL_REQUEST ${GOPATH}/bin/goveralls -coverprofile=./bin/cover.out -repotoken $C_TOKEN; \
    else \
        export CGO_ENABLED=0 && \
        GOOS=linux GOARCH=amd64 GOARM= BINTRAY_API_KEY=${BINTRAY_API_KEY} BINTRAY_API_USER=${BINTRAY_API_USER} go run cmd/bintray/upload.go /app/plugins ${VERSION} ${GOARCH}; \
    fi;

RUN mkdir -p /out && \
    cp /app/go-home /out/ && \
    if [ "x${GOARM}" != "x" ]; then \
          curl -L -o qemu-arm-static.tar.gz ${QEMU} && \
          tar xzf qemu-arm-static.tar.gz && \
          cp qemu-arm-static /out/; \
    fi;

##################################################################################################

FROM $RUN_IMAGE

COPY --from=build /out/* /usr/bin/

CMD ["go-home"]