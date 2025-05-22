COVERAGE_FILE := cover.out
PKG := ./...

.PHONY: all build run test cover fmt lint mocks clean

all: build

build:
	go build -o cmd/agent/agent ./cmd/agent/
	go build -o cmd/server/server ./cmd/server/

build-server:
	go build -o cmd/server/server ./cmd/server/

build-agent:
	go build -o cmd/agent/agent ./cmd/agent/

test:
	go test ./... -v

cover:
	go test -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -func=$(COVERAGE_FILE)

cover-html:
	go test -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -html=$(COVERAGE_FILE)