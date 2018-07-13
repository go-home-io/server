# Go params
GOCMD=PATH=${PATH}:$(GO_BIN_FOLDER) go
GOGET=$(GOCMD) get
GOBUILD=$(GOCMD) build
GOGENERATE=$(GOCMD) generate

GO_BIN_FOLDER=$(GOPATH)/bin
METALINER=$(GO_BIN_FOLDER)/gometalinter.v2

PLUGINS_LOCATION=$(GOPATH)/src/github.com/go-home-io/providers/
BIN_FOLDER=${CURDIR}/bin
BIN_NAME=$(BIN_FOLDER)/go-home
PLUGINS_BINS=${CURDIR}/bin/plugins

.PHONY: build dep lint generate build-plugins run-server test-local
define build_plugins_task =
	rm -rf $(PLUGINS_BINS)
	cd $(PLUGINS_LOCATION)
	for plugin_type in *; do
		if [ -d "$${plugin_type}" ]; then
			for plugin in $${plugin_type}/*; do
				if [ -d "$${plugin}" ]; then
					cd $${plugin}
					echo "Building $${plugin}"
					$(GOBUILD) -buildmode=plugin -o $(PLUGINS_BINS)/$${plugin}.so .
					cd $(PLUGINS_LOCATION)
				fi;
			done;
		fi;
	done;
endef

build:
	$(GOBUILD) -o $(BIN_NAME) -v

dep:
	$(GOGET) github.com/alecthomas/gometalinter
	$(METALINER) --install

lint:
	$(METALINER) ./...

generate:
	$(GOGENERATE) -v ./...
	@cd $(PLUGINS_LOCATION)
	$(GOGENERATE) -v ./...

run-server: build
	$(BIN_NAME) --configs=${CURDIR}/config --plugins=${CURDIR}/bin/plugins

run-worker: build
	$(BIN_NAME) -c provider:fs -c location:${CURDIR}/configs -p ${CURDIR}/bin/plugins -w

test:
	$(GOCMD) test --covermode=count -coverprofile=$(BIN_FOLDER)/cover.out.tmp ./...
	@cat $(BIN_FOLDER)/cover.out.tmp | grep -v "fake_" | grep -v "_enumer" | grep -v "mocks" > $(BIN_FOLDER)/cover.out
	$(GOCMD) tool cover --html=$(BIN_FOLDER)/cover.out

test-local: test
	@rm -f $(BIN_FOLDER)/cover.out.tmp

.ONESHELL:
SHELL = /bin/bash
build-plugins:
	$(build_plugins_task)