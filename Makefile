# Go params
GO_BIN_FOLDER=$(GOPATH)/bin
GOCMD=GOARM=${GOARM} GOARCH=${GOARCH} PATH=${PATH}:$(GO_BIN_FOLDER) GO111MODULE=on go

GOGET=$(GOCMD) get
GOGET_NO_MOD=GOARM=${GOARM} GOARCH=${GOARCH} PATH=${PATH}:$(GO_BIN_FOLDER) GO111MODULE=off go get
GOBUILD=$(GOCMD) build -ldflags="-s -w"
GOGENERATE=$(GOCMD) generate

MOD=$(GOCMD) mod tidy
MOD_RESTORE=$(GOGET) -v -d ./... && $(GOCMD) mod vendor
GOVERALLS=$(GO_BIN_FOLDER)/goveralls

PLUGINS_LOCATION=$(GOPATH)/src/github.com/go-home-io/providers/
BIN_FOLDER=${CURDIR}/bin
BIN_NAME=$(BIN_FOLDER)/go-home
PLUGINS_BINS=$(BIN_FOLDER)/plugins

#METALINER=$(GO_BIN_FOLDER)/gometalinter
METALINER=PATH=${PATH}:$(BIN_FOLDER) $(BIN_FOLDER)/gometalinter

.PHONY: utilities-build utilities-ci utilities build-server build-plugins build run-server run-worker test-local test

define build_plugins_task =
	set -e
	rm -rf $(PLUGINS_BINS)
	cd $(PLUGINS_LOCATION)
	for plugin_type in *; do
		if [ -d "$${plugin_type}" ]; then
			for plugin in $${plugin_type}/*; do
				if [ -d "$${plugin}" ]; then
					cd $${plugin}
					echo "======================================="
					echo "Building $${plugin}"
					echo "======================================="
					$(GOBUILD) -buildmode=plugin -o $(PLUGINS_BINS)/$${plugin}.so .
					cd $(PLUGINS_LOCATION)
				fi;
			done;
		fi;
	done;
endef

define validate_dependencies =
	set -e
	echo "======================================="
	echo "Validating server"
	echo "======================================="
	$(MOD)
	cd $(PLUGINS_LOCATION)
	for plugin_type in *; do
		if [ -d "$${plugin_type}" ]; then
			for plugin in $${plugin_type}/*; do
				if [ -d "$${plugin}" ]; then
					cd $${plugin}
					echo "======================================="
					echo "Validating $${plugin}"
					echo "======================================="
					$(MOD)
					cd $(PLUGINS_LOCATION)
				fi;
			done;
		fi;
	done;
endef

define restore_dependencies =
	set -e
	echo "======================================="
	echo "Validating server"
	echo "======================================="
	$(MOD_RESTORE)
	cd $(PLUGINS_LOCATION)
	for plugin_type in *; do
		if [ -d "$${plugin_type}" ]; then
			for plugin in $${plugin_type}/*; do
				if [ -d "$${plugin}" ]; then
					cd $${plugin}
					echo "======================================="
					echo "Validating $${plugin}"
					echo "======================================="
					$(MOD_RESTORE)
					cd $(PLUGINS_LOCATION)
				fi;
			done;
		fi;
	done;
endef

define lint_all =
	set -e
	echo "======================================="
	echo "Linting server"
	echo "======================================="
	$(METALINER) --enable=megacheck --config=${CURDIR}/.gometalinter.json --sort=linter ./...
	cd $(PLUGINS_LOCATION)
	for plugin_type in *; do
    		if [ -d "$${plugin_type}" ]; then
    			for plugin in $${plugin_type}/*; do
    				if [ -d "$${plugin}" ]; then
    					echo "======================================="
    					echo "Linting $${plugin}"
    					echo "======================================="
    					cd $${plugin}
    					$(METALINER) --config=${CURDIR}/.gometalinter.json --enable=megacheck_provider --sort=linter ./...
    					cd $(PLUGINS_LOCATION)
    				fi;
    			done;
    		fi;
    	done;
endef

utilities-build:
	$(GOGET) github.com/alvaroloes/enumer
	$(GOGET) github.com/rakyll/statik

utilities-ci:
	#$(GOGET_NO_MOD) github.com/alecthomas/gometalinter
	#$(METALINER) --install --update
	curl -L https://git.io/vp6lP | sh
	$(GOGET) github.com/mattn/goveralls

utilities: utilities-build utilities-ci

build-server:
	$(GOBUILD) -ldflags "-X github.com/go-home-io/server/utils.Version=${VERSION} \
		-X github.com/go-home-io/server/utils.Arch=${GOARCH}" -o $(BIN_NAME)

build: build-plugins build-server

generate:
	$(GOGENERATE) -v ./...
	@cd $(PLUGINS_LOCATION)
	$(GOGENERATE) -v ./...

run-server:
	$(BIN_NAME) -c provider:fs -c location:${CURDIR}/configs -p ${CURDIR}/bin/plugins

run-only-worker:
	$(BIN_NAME) -c provider:fs -c location:${CURDIR}/configs -p ${CURDIR}/bin/plugins -w

run-worker: build run-only-worker

test:
	@set -e
	$(GOCMD) test -failfast --covermode=count -coverprofile=$(BIN_FOLDER)/cover.out.tmp ./...
	@cat $(BIN_FOLDER)/cover.out.tmp | grep -v "fake_" | grep -v "_enumer" | grep -v "mocks" | grep -v "public" | grep -v "statik" | grep -v "server/api_" | grep -v "cmd/" > $(BIN_FOLDER)/cover.out
	@rm -f $(BIN_FOLDER)/cover.out.tmp

test-local: test
	$(GOCMD) tool cover --html=$(BIN_FOLDER)/cover.out

git: dep-ensure generate lint test-local

build-rpi-cache-docker:
	docker build -t go-home-cahe -f Dockerfile.rpi.cache .

.ONESHELL:
SHELL = /bin/sh
build-plugins:
	$(build_plugins_task)

.ONESHELL:
SHELL = /bin/sh
dep:
	$(restore_dependencies)

.ONESHELL:
SHELL = /bin/sh
lint:
	$(lint_all)

.ONESHELL:
dep-ensure:
	$(validate_dependencies)