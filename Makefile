HOSTNAME = terraform.example.com
NAMESPACE = local
NAME = looker

BINARY = terraform-provider-${NAME}
VERSION ?= 0.0.1
BUILD_DIR ?= $(CURDIR)/out

GO_PACKAGES := $(shell go list ./... | grep -v vendor) 
GO_FILES := $(shell find . -name '*.go' | grep -v vendor)
GO_OS ?= $(shell go env GOOS)
GO_ARCH ?= $(shell go env GOARCH)

.PHONY: build
build:
	@go build -v -o ${BUILD_DIR}/${BINARY}_v${VERSION}

.PHONY: install
install: build
	@mkdir -p ${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/$(VERSION)/${GO_OS}_${GO_ARCH}
	@cp ${BUILD_DIR}/${BINARY}_v${VERSION} ${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/$(VERSION)/${GO_OS}_${GO_ARCH}

.PHONY: gen
gen:
	@go generate ./...

.PHONY: clean
clean: 
	@rm -rf ${BUILD_DIR} ${HOME}/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}
	@rm -rf ${NAMESPACE}/.terraform
	@rm -rf ${NAMESPACE}/.terraform.lock.hcl

.PHONY: lint
lint:
	@golangci-lint run -c .golangci.yaml

.PHONY: test
test: 
	@go test ${GO_PACKAGES} || exit 1                                                   
	@echo ${GO_PACKAGES} | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4                    

.PHONY: testacc
testacc:
	@TF_ACC=1 go test ${GO_PACKAGES} -v $(TESTARGS) -timeout 120m

.PHONY: rectestacc
rectestacc:
	@TF_ACC=1 TF_REC=true go test ${GO_PACKAGES} -v $(TESTARGS) -timeout 120m

.PHONY: sweep
sweep:
	@echo "WARNING: This will destroy infrastructure. Only use on development instances."
	@go test ./looker -v -sweep="phony" $(SWEEPARGS) -timeout 60m

local: clean install
	@cd $(NAMESPACE) && terraform init
