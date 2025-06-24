COVERAGE_FILE := cover.out
NO_MOCKS_COVERAGE_FILE := clean_cover.out
PKG := ./...
SERVER_PATH := ./cmd/server
AGENT_PATH := ./cmd/agent

BUILD_VERSION := v0.0.20
DATE := $(shell date -u +"%d %b %y %H:%M %z")
COMMIT := $(shell git log --pretty=format:%Creset%s --no-merges -1)
ldflags := -ldflags="-X 'main.buildVersion=$(BUILD_VERSION)' -X 'main.buildDate=$(DATE)' -X 'main.buildCommit=$(COMMIT)'"

.PHONY: all build run test cover fmt lint mocks clean

all: build

build:
	go build -o $(AGENT_PATH)/agent $(.AGENT_PATH)
	go build -o $(SERVER_PATH)/server $(SERVER_PATH)

build-server:
	go build -o $(SERVER_PATH)/server $(SERVER_PATH)

build-agent:
	go build -o $(AGENT_PATH)/agent $(.AGENT_PATH)

build-staticlint:
	go build -o cmd/staticlint/staticlint ./cmd/staticlint/

release-server:
	go build $(ldflags) -o $(SERVER_PATH)/server $(SERVER_PATH)

release-agent:
	go build $(ldflags) -o $(AGENT_PATH)/agent $(AGENT_PATH)

staticlint:
	go vet -vettool=$(which statictest) ./...
	go vet -vettool=./cmd/staticlint/staticlint ./...

betteralign:
	betteralign -apply -test_files ./...

test:
	go test ./... -v

cover:
	go test -coverprofile=$(COVERAGE_FILE) ./...
	grep -v "internal/mocks" $(COVERAGE_FILE) > $(NO_MOCKS_COVERAGE_FILE)
	# grep -Ev "internal/mocks|memstorage|postgres" $(COVERAGE_FILE) > $(NO_MOCKS_COVERAGE_FILE)
	go tool cover -func=$(NO_MOCKS_COVERAGE_FILE)

cover-ugly:
	go test -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -func=$(COVERAGE_FILE)

cover-html:
	go test -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -html=$(COVERAGE_FILE)
