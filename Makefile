VERSION := $(shell git describe --tags 2>/dev/null || git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S%z')
GIT_COMMIT := $(shell git rev-parse --short HEAD)

META := github.com/fabedge/fabctl/pkg/about
FLAG_VERSION := ${META}.version=${VERSION}
FLAG_BUILD_TIME := ${META}.buildTime=${BUILD_TIME}
FLAG_GIT_COMMIT := ${META}.gitCommit=${GIT_COMMIT}
GOLDFLAGS ?= -s -w
LDFLAGS := -ldflags "${GOLDFLAGS} -X ${FLAG_VERSION} -X ${FLAG_BUILD_TIME} -X ${FLAG_GIT_COMMIT}"

OUTPUT_DIR := _output

fmt:
	GOOS=linux go fmt ./...

vet:
	GOOS=linux go vet ./...

.PHONY: fabctl
fabctl:
	go build ${LDFLAGS} -o ${OUTPUT_DIR}/$@ .