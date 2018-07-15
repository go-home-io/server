# Go params
GO_BIN_FOLDER=$(GOPATH)/bin
GOCMD=go

GOGET=$(GOCMD) get
GOBUILD=$(GOCMD) build
GOGENERATE=$(GOCMD) generate

METALINER=$(GO_BIN_FOLDER)/gometalinter.v2
DEP=$(GO_BIN_FOLDER)/dep ensure
GLIDE=$(GO_BIN_FOLDER)/glide install

PLUGINS_LOCATION=$(GOPATH)/src/github.com/go-home-io/providers/
BIN_FOLDER=${CURDIR}/bin
BIN_NAME=$(BIN_FOLDER)/go-home
PLUGINS_BINS=$(BIN_FOLDER)/plugins

.PHONY: build dep lint generate build-plugins run-server test-local

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

define restore_dependencies =
	set -e
	echo "======================================="
	echo "Restoring server"
	echo "======================================="
	$(DEP)
	cd $(PLUGINS_LOCATION)
	for plugin_type in *; do
		if [ -d "$${plugin_type}" ]; then
			for plugin in $${plugin_type}/*; do
				if [ -d "$${plugin}" ]; then
					cd $${plugin}
					echo "======================================="
					echo "Restoring $${plugin}"
					echo "======================================="
					if [ -f ./glide.yaml ]; then
						$(GLIDE)
					else
						$(DEP)
					fi
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
	$(METALINER) --enable=megacheck --sort=linter ./...
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

utilities:
	$(GOGET) github.com/Masterminds/glide
	$(GOGET) gopkg.in/alecthomas/gometalinter.v2
	$(GOGET) github.com/golang/dep/cmd/dep
	$(METALINER) --install

build:
	$(GOBUILD) -o $(BIN_NAME) -v

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
	@rm -f $(BIN_FOLDER)/cover.out.tmp

test-local: test
	$(GOCMD) tool cover --html=$(BIN_FOLDER)/cover.out

.ONESHELL:
SHELL = /bin/bash
build-plugins:
	$(build_plugins_task)

.ONESHELL:
SHELL = /bin/bash
dep:
	$(restore_dependencies)

.ONESHELL:
SHELL = /bin/bash
lint:
	$(lint_all)