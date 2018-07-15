ARG BUILD_IMAGE
ARG RUN_IMAGE
FROM $BUILD_IMAGE as build

ENV PROVIDERS=https://github.com/go-home-io/providers.git \
    HOME_DIR=${GOPATH}/src/github.com/go-home-io/server

WORKDIR ${HOME_DIR}

COPY . .

ARG INSTALL_LIBS
RUN /bin/sh -c "${INSTALL_LIBS}" && \
    make utilities-build && \
    cd ${GOPATH} && \
    mkdir -p src/github.com/go-home-io && \
    cd src/github.com/go-home-io && \
    git clone ${PROVIDERS} && \
    cd ${HOME_DIR} && \
    make dep

RUN mkdir -p /app && \
    make BIN_FOLDER=/app build

ARG LINT
ARG C_TOKEN
RUN if [ "${LINT}" != "false" ]; then \
        set -e && \
        mkdir bin && \
        make utilities-ci && \
        make lint && \
        make test && \
        ${GOPATH}/bin/goveralls -coverprofile=./bin/cover.out -repotoken $C_TOKEN; \
    fi;

##################################################################################################

FROM $RUN_IMAGE

ENV HOME_DIR=/go-home

WORKDIR ${HOME_DIR}

COPY --from=build /app/ ./

CMD ["./go-home"]