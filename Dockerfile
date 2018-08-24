ARG BUILD_IMAGE
ARG RUN_IMAGE
ARG NODE_IMAGE
FROM $NODE_IMAGE as node

ENV UI=https://github.com/nicknesk/gohome-prototype-react

WORKDIR /

RUN apk update && apk add git && \
    git clone ${UI} dashboard && \
    cd dashboard && \
    npm i && \
    npm run build

##################################################################################################

FROM $BUILD_IMAGE as build


ENV PROVIDERS=https://github.com/go-home-io/providers.git \
    HOME_DIR=${GOPATH}/src/github.com/go-home-io/server

WORKDIR ${HOME_DIR}

COPY . .
COPY --from=node /dashboard/build/* ./public/

RUN apk update && apk add make git gcc libc-dev ca-certificates && \
    make utilities-build && \
    cd ${GOPATH} && \
    mkdir -p src/github.com/go-home-io && \
    cd src/github.com/go-home-io && \
    git clone ${PROVIDERS} && \
    cd ${HOME_DIR}

ARG GOARCH
ENV GOARCH=${GOARCH}

ARG TRAVIS_TAG
ENV VERSION=${TRAVIS_TAG}

ARG GOARM
RUN mkdir -p /app && \
    VERSION=${VERSION} GOARM=${GOARM} GOARCH=${GOARCH} make generate && \
    VERSION=${VERSION} GOARM=${GOARM} GOARCH=${GOARCH} make BIN_FOLDER=/app build

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
        mkdir bin && \
        make utilities-ci && \
        make lint && \
        make test && \
        TRAVIS=$TRAVIS TRAVIS_JOB_ID=$TRAVIS_JOB_ID TRAVIS_BRANCH=$TRAVIS_BRANCH TRAVIS_PULL_REQUEST=$TRAVIS_PULL_REQUEST ${GOPATH}/bin/goveralls -coverprofile=./bin/cover.out -repotoken $C_TOKEN; \
    else \
        BINTRAY_API_KEY=${BINTRAY_API_KEY} BINTRAY_API_USER=${BINTRAY_API_USER} go run build/main.go /app/plugins ${VERSION} ${GOARCH}; \
    fi;

##################################################################################################

FROM $RUN_IMAGE

ENV HOME_DIR=/go-home

WORKDIR ${HOME_DIR}

RUN apk update && apk add ca-certificates

COPY --from=build /app/go-home .

CMD ["./go-home"]