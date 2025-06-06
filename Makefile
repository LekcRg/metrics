COVERAGE_FILE := cover.out
NO_MOCKS_COVERAGE_FILE := clean_cover.out
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

build-staticlint:
	go build -o cmd/staticlint/staticlint ./cmd/staticlint/

staticlint:
	go vet -vettool=./cmd/staticlint/staticlint ./...

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
