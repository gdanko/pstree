GOPATH := $(shell go env GOPATH)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
PSTREE_VERSION := 0.6.2

GOOS ?= $(shell uname | tr '[:upper:]' '[:lower:]')
GOARCH ?=$(shell arch)

.PHONY: all build install

all: build install

.PHONY: mod-tidy
mod-tidy:
	go mod tidy

.PHONY: build OS ARCH
build: guard-PSTREE_VERSION mod-tidy clean
	@echo "================================================="
	@echo "Building pstree"
	@echo "=================================================\n"

	@if [ ! -d "bin" ]; then \
		mkdir "bin"; \
	fi
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o "bin/pstree"
	sleep 2
	tar -czvf "pstree_${PSTREE_VERSION}_${GOOS}_${GOARCH}.tgz" bin; \

.PHONY: clean
clean:
	@echo "================================================="
	@echo "Cleaning pstree"
	@echo "=================================================\n"
	@if [ -f bin/pstree ]; then \
		rm -f bin/pstree; \
	fi; \

.PHONY: clean-all
clean-all: clean
	@echo "================================================="
	@echo "Cleaning tarballs"
	@echo "=================================================\n"
	@rm -f *.tgz 2>/dev/null

.PHONY: install
install:
	@echo "================================================="
	@echo "Installing pstree in ${GOPATH}/bin"
	@echo "=================================================\n"

	go install -race

#
# General targets
#
guard-%:
	@if [ "${${*}}" = "" ]; then \
		echo "Environment variable $* not set"; \
		exit 1; \
	fi
