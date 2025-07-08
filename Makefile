OUTPUT_DIR = .go-builds
VERSION ?= local

.PHONY: clean build-proto build-all \
	build-linux-amd64 build-linux-arm64 \
	build-darwin-amd64 build-darwin-arm64 \
	build-windows-amd64

# GO
clean:
	rm -rf $(OUTPUT_DIR) || true

build-proto:
	@command -v protoc >/dev/null 2>&1 || { echo >&2 "protoc not installed. Aborting."; exit 1; }
	protoc --go_out=. --go-grpc_out=. ./proto/*.proto

build-all: clean build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

build-linux-amd64: build-proto
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" \
		-o $(OUTPUT_DIR)/kubensage-agent-$(VERSION)-linux-amd64 cmd/kubensage-agent/main.go

build-linux-arm64: build-proto
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" \
		-o $(OUTPUT_DIR)/kubensage-agent-$(VERSION)-linux-arm64 cmd/kubensage-agent/main.go

build-darwin-amd64: build-proto
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" \
		-o $(OUTPUT_DIR)/kubensage-agent-$(VERSION)-darwin-amd64 cmd/kubensage-agent/main.go

build-darwin-arm64: build-proto
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" \
		-o $(OUTPUT_DIR)/kubensage-agent-$(VERSION)-darwin-arm64 cmd/kubensage-agent/main.go

build-windows-amd64: build-proto
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" \
		-o $(OUTPUT_DIR)/kubensage-agent-$(VERSION)-windows-amd64.exe cmd/kubensage-agent/main.go

fresh-scp: build-linux-amd64
	scp .go-builds/kubensage-agent-local-linux-amd64 roman@192.168.1.160:/home/roman/kubensage/agent