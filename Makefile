VERSION ?=
GIT_SHA ?= $(shell git rev-parse --short HEAD 2>/dev/null)
GIT_TAG ?= $(shell git describe --tags --abbrev=0 2>/dev/null)

ifeq ($(VERSION),)
  ifneq ($(GIT_TAG),)
    VERSION := $(GIT_TAG)
  else ifneq ($(GIT_SHA),)
    VERSION := $(GIT_SHA)
  else
    VERSION := unknown
  endif
endif

LDFLAGS := -s -w -X github.com/nextnowlabs/vision-cli/internal/cli.Version=$(VERSION)

.PHONY: build install clean test lint fmt

build:
	go build -ldflags "$(LDFLAGS)" -o bin/vg ./cmd/vg/

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/vg/

clean:
	rm -rf bin/

test:
	go test ./...

lint:
	go vet ./...

fmt:
	go fmt ./...
