-include config.mk

OUTPUT_DIR = build
MODULE := github.com/kubensage/kubensage-agent
VERSION ?= local

.PHONY: build-proto \
		vet clean tidy build build-linux-amd64 build-linux-arm64 \
		fresh-scp

# Proto
build-proto:
	@command -v protoc >/dev/null 2>&1 || { echo >&2 "protoc not installed. Aborting."; exit 1; }
	protoc --go_out=. --go-grpc_out=. ./proto/*.proto

# Go
clean:
	rm -rf $(OUTPUT_DIR) || true

tidy:
	go mod tidy

vet:
	go vet ./...

build-linux-amd64: vet build-proto tidy
	GOOS=linux GOARCH=amd64 go build -ldflags "-X '$(MODULE)/pkg/buildinfo.Version=$(VERSION)'" \
		-o $(OUTPUT_DIR)/kubensage-agent-$(VERSION)-linux-amd64 cmd/kubensage-agent/main.go

build-linux-arm64: vet build-proto tidy
	GOOS=linux GOARCH=arm64 go build -ldflags "-X '$(MODULE)/pkg/buildinfo.Version=$(VERSION)'" \
		-o $(OUTPUT_DIR)/kubensage-agent-$(VERSION)-linux-arm64 cmd/kubensage-agent/main.go

build: clean build-linux-amd64 build-linux-arm64

# Utils
fresh-scp: build-linux-amd64
	scp $(OUTPUT_DIR)/kubensage-agent-$(VERSION)-linux-amd64 $(REMOTE_USER)@$(REMOTE_HOST):$(REMOTE_PATH)