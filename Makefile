VERSION  ?= 1.8.12
BINARY   := serve
IMAGE    := nhoque/static-file-server

LDFLAGS  := -s -w -X github.com/N-Hoque/static-file-server/pkg/cli/version.version=$(VERSION)

.DEFAULT_GOAL := help

.PHONY: build test lint coverage cross-build docker-build clean help

build: ## Compile the binary for the current platform
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

test: ## Run unit tests
	go test ./...

lint: ## Run golangci-lint
	golangci-lint run

coverage: ## Run tests and produce coverage.out + coverage.html
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

cross-build: ## Cross-compile for all supported platforms into ./out
	@mkdir -p out
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64         go build -ldflags "$(LDFLAGS)" -o out/$(BINARY)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64          go build -ldflags "$(LDFLAGS)" -o out/$(BINARY)-linux-arm64 .
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm   GOARM=7  go build -ldflags "$(LDFLAGS)" -o out/$(BINARY)-linux-arm7 .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64          go build -ldflags "$(LDFLAGS)" -o out/$(BINARY)-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64          go build -ldflags "$(LDFLAGS)" -o out/$(BINARY)-darwin-arm64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64          go build -ldflags "$(LDFLAGS)" -o out/$(BINARY)-windows-amd64.exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64          go build -ldflags "$(LDFLAGS)" -o out/$(BINARY)-windows-arm64.exe .

docker-build: ## Build the Docker image for the current platform
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE):$(VERSION) .

clean: ## Remove build artefacts
	rm -f $(BINARY)
	rm -rf out/
	rm -f coverage.out coverage.html

help: ## Show this help
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "  %-14s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
