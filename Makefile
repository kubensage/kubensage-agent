OUTPUT_DIR = .go-builds
VERSION ?= 1.0.0

.PHONY: clean build-all \
	build-linux-amd64 build-linux-arm64 \
	build-darwin-amd64 build-darwin-arm64 \
	build-windows-amd64 tag

# GO
clean:
	rm -rf $(OUTPUT_DIR) || true

build-all: clean build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" \
		-o $(OUTPUT_DIR)/kubensage-agent-linux-amd64 cmd/kubensage-agent/main.go

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" \
		-o $(OUTPUT_DIR)/kubensage-agent-linux-arm64 cmd/kubensage-agent/main.go

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" \
		-o $(OUTPUT_DIR)/kubensage-agent-darwin-amd64 cmd/kubensage-agent/main.go

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" \
		-o $(OUTPUT_DIR)/kubensage-agent-darwin-arm64 cmd/kubensage-agent/main.go

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" \
		-o $(OUTPUT_DIR)/kubensage-agent-windows-amd64.exe cmd/kubensage-agent/main.go

# GIT
tag:
ifndef TAG
	$(error TAG is not set. Usage: make tag TAG=v1.0.0)
endif
	git tag $(TAG) -m "Release $(TAG)"
	git push origin $(TAG)