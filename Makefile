VERSION ?= 1.0.0

.PHONY: clean build-all \
	build-linux-amd64 build-linux-arm64 \
	build-darwin-amd64 build-darwin-arm64 \
	build-windows-amd64

clean:
	rm -rf go-builds

build-all: clean build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" \
		-o go-builds/kubensage-agent-linux-amd64 .

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" \
		-o go-builds/kubensage-agent-linux-arm64 .

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" \
		-o go-builds/kubensage-agent-darwin-amd64 .

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" \
		-o go-builds/kubensage-agent-darwin-arm64 .

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" \
		-o go-builds/kubensage-agent-windows-amd64.exe .