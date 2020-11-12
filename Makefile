.PHONY: none
none:


deps: go.mod
	@echo "+ $@"
	@go mod tidy
	@go mod download
	@go mod verify
	@touch deps

UNAME_S := $(shell uname -s)
HOST_OS := linux
ifeq ($(UNAME_S),Darwin)
    HOST_OS := darwin
endif

TAG := $(shell ./get-tag)

GOBIN := $(CURDIR)/.gobin
PATH := $(GOBIN):$(PATH)
# Makefile on Mac doesn't pass this updated PATH to the shell
# and so without the following line, the shell does not end up
# trying commands in $(GOBIN) first.
# See https://stackoverflow.com/a/36226784/3690207
SHELL := env PATH=$(PATH) /bin/bash

########################################
###### Binaries we depend on ###########
########################################

GOLANGCILINT_BIN := $(GOBIN)/golangci-lint
$(GOLANGCILINT_BIN): deps
	@echo "+ $@"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint

STATICCHECK_BIN := $(GOBIN)/staticcheck
$(STATICCHECK_BIN): deps
	@echo "+ $@"
	@go install honnef.co/go/tools/cmd/staticcheck

###########
## Lint  ##
###########

.PHONY: golangci-lint
golangci-lint: $(GOLANGCILINT_BIN)
ifdef CI
	@echo '+ $@'
	@echo 'The environment indicates we are in CI; running linters in check mode.'
	@echo 'If this fails, run `make lint`.'
	golangci-lint run
else
	golangci-lint run --fix
endif

.PHONY: staticcheck
staticcheck: $(STATICCHECK_BIN)
	staticcheck -checks=all,-ST1000 ./...

.PHONY: lint
lint: golangci-lint staticcheck


##########
## Test ##
##########

.PHONY: test
test: 
	go test ./...


###########
## Build ##
###########

.PHONY: tag
tag:
	@echo $(TAG)

.PHONY: build
build:
	@echo "+ $@"
	@mkdir -p "${GOBIN}"
	@CGO_ENABLED=0 GOOS=linux \
	go build -a -ldflags "-s -w" \
		-o ${GOBIN}/linux/network-crawler ./cmd/network-crawler
	@CGO_ENABLED=0 GOOS=darwin \
	go build -a -ldflags "-s -w" \
		-o ${GOBIN}/darwin/network-crawler ./cmd/network-crawler
	@cp ${GOBIN}/${HOST_OS}/network-crawler ${GOBIN}/network-crawler

.PHONY: image
image: build
	@echo "+ $@"
	@cp ${GOBIN}/linux/network-crawler ./image
	@docker build -t us.gcr.io/stackrox-hub/network-crawler:$(TAG) image

.PHONY: push
push: image
	@echo "+ $@"
	@docker push us.gcr.io/stackrox-hub/network-crawler:$(TAG) | cat
