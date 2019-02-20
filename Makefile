SHELL := /bin/bash
GO := GO15VENDOREXPERIMENT=1 go
NAME := locations
OS := $(shell uname)
MAIN_GO := main.go
ROOT_PACKAGE := $(GIT_PROVIDER)/$(ORG)/$(NAME)
GO_VERSION := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
PACKAGE_DIRS := $(shell $(GO) list ./... | grep -v /vendor/)
PKGS := $(shell go list ./... | grep -v /vendor | grep -v generated)
PKGS := $(subst  :,_,$(PKGS))
BUILDFLAGS := ''
CGO_ENABLED = 0
VENDOR_DIR=vendor
PROTO_PATH="api/proto"
JS_OUT_DIR="api/js"
GO_OUT_DIR="api/go"
GRPC_WEB_OUT_DIR="api/web"

all: build

check: fmt build test

get-deps:
	wget https://github.com/grpc/grpc-web/releases/download/1.0.3/protoc-gen-grpc-web-1.0.3-linux-x86_64
	chmod +x protoc-gen-grpc-web-1.0.3-linux-x86_64
	mv protoc-gen-grpc-web-1.0.3-linux-x86_64 /usr/bin/protoc-gen-grpc-web
	git clone https://github.com/golang/protobuf
	cd protobuf/protoc-gen-go; git checkout tags/v1.2.0 -b v1.2.0
	cd protobuf/protoc-gen-go; go build
	cp protobuf/protoc-gen-go/protoc-gen-go /usr/bin/protoc-gen-go
	cd protobuf/protoc-gen-go; go install
	rm -rf protobuf

proto:
	mkdir -p $(GO_OUT_DIR)
	mkdir -p $(JS_OUT_DIR)
	mkdir -p $(GRPC_WEB_OUT_DIR)
	protoc \
        --proto_path=${PROTO_PATH}:. \
        --go_out=plugins=grpc:${GO_OUT_DIR} \
        --grpc-web_out=import_style=commonjs,mode=grpcwebtext:${GRPC_WEB_OUT_DIR} \
        --js_out="import_style=commonjs,binary:${JS_OUT_DIR}" \
        $(PROTO_PATH)/locations.proto

build:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -ldflags $(BUILDFLAGS) -o bin/$(NAME) $(MAIN_GO)

test: 
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(PACKAGE_DIRS) -test.v

full: $(PKGS)

install:
	GOBIN=${GOPATH}/bin $(GO) install -ldflags $(BUILDFLAGS) $(MAIN_GO)

fmt:
	@FORMATTED=`$(GO) fmt $(PACKAGE_DIRS)`
	@([[ ! -z "$(FORMATTED)" ]] && printf "Fixed unformatted files:\n$(FORMATTED)") || true

clean:
	rm -rf build release
	rm -rf $(JS_OUT_DIR)
	rm -rf $(GO_OUT_DIR)
	rm -rf $(GRPC_WEB_OUT_DIR)
	rm -rf protobuf

linux:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GO) build -ldflags $(BUILDFLAGS) -o bin/$(NAME) $(MAIN_GO)

.PHONY: release clean

FGT := $(GOPATH)/bin/fgt
$(FGT):
	go get github.com/GeertJohan/fgt

GOLINT := $(GOPATH)/bin/golint
$(GOLINT):
	go get github.com/golang/lint/golint

$(PKGS): $(GOLINT) $(FGT)
	@echo "LINTING"
	@$(FGT) $(GOLINT) $(GOPATH)/src/$@/*.go
	@echo "VETTING"
	@go vet -v $@
	@echo "TESTING"
	@go test -v $@

.PHONY: lint
lint: vendor | $(PKGS) $(GOLINT) # ❷
	@cd $(BASE) && ret=0 && for pkg in $(PKGS); do \
	    test -z "$$($(GOLINT) $$pkg | tee /dev/stderr)" || ret=1 ; \
	done ; exit $$ret

watch:
	reflex -r "\.go$" -R "vendor.*" make skaffold-run

skaffold-run: build
	skaffold run -p dev
