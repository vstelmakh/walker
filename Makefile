.DEFAULT_GOAL := help
.SILENT:
.PHONY: help test vet fmt get tidy clean build build-all build-macos-arm64 build-macos-amd64 build-linux-amd64 build-linux-arm64 build-windows-amd64

# Go version (read from go.mod)
GO_VERSION := $(shell awk '/^go /{print $$2}' go.mod)
GO_IMAGE ?= golang:$(GO_VERSION)-alpine

# Docker command to run Go as the host user, so files created in the project are not owned by root.
# Module and build caches are kept in one Docker volume, mounted at /go because that directory is writable for any user.
DOCKER_GO = docker run --rm \
	--user "$$(id -u):$$(id -g)" \
	-e GOCACHE=/go/build-cache \
	-v "$(CURDIR)":/app -w /app \
	-v go-cache:/go

# Current time in milliseconds, uses perl because date on macOS has no sub-second precision
NOW_MS = perl -MTime::HiRes=time -e 'printf "%d", time * 1000'

help: ## Show available commands
	grep -hE '^[a-zA-Z][a-zA-Z0-9-]*:.*## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*## "}; {printf "  \033[32m%-24s\033[0m %s\n", $$1, $$2}'

test: ## Run tests
	$(DOCKER_GO) $(GO_IMAGE) go test ./...

vet: ## Report suspicious Go code
	$(DOCKER_GO) $(GO_IMAGE) go vet ./...

fmt: ## Format Go files
	$(DOCKER_GO) $(GO_IMAGE) gofmt -l -w .

# Usage example: make get PKG=github.com/foo/bar@latest
get: ## Add or update a Go module, requires PKG argument to be set
	[ -n "$(PKG)" ] || { echo 'PKG is required, example: make get PKG=github.com/foo/bar@latest' >&2; exit 1; }
	$(DOCKER_GO) $(GO_IMAGE) go get $(PKG)

tidy: ## Add missing and remove unused Go modules
	$(DOCKER_GO) $(GO_IMAGE) go mod tidy

clean: ## Remove build output
	rm -rf dist

build-macos-arm64: ## Build for macOS, Apple Silicon
	$(MAKE) build GOOS=darwin GOARCH=arm64

build-macos-amd64: ## Build for macOS, Intel
	$(MAKE) build GOOS=darwin GOARCH=amd64

build-linux-amd64: ## Build for Linux, x86-64
	$(MAKE) build GOOS=linux GOARCH=amd64

build-linux-arm64: ## Build for Linux, ARM64
	$(MAKE) build GOOS=linux GOARCH=arm64

build-windows-amd64: ## Build for Windows, x86-64
	$(MAKE) build GOOS=windows GOARCH=amd64

# Usage example: make build GOOS=linux GOARCH=amd64
# Binary file name says "macos", because Go calls macOS "darwin"
build: ## Build binary, requires GOOS and GOARCH arguments to be set
	[ -n "$(GOOS)" ] && [ -n "$(GOARCH)" ] || { echo 'GOOS and GOARCH are required, example: make build GOOS=linux GOARCH=amd64' >&2; exit 1; }
	mkdir -p dist
	printf 'Platform: %s/%s\n' '$(GOOS)' '$(GOARCH)'
	OUT=dist/walker-$(if $(filter darwin,$(GOOS)),macos,$(GOOS))-$(GOARCH)$(if $(filter windows,$(GOOS)),.exe); \
	printf 'Binary:   %s\n' "$$OUT"; \
	START=$$($(NOW_MS)); \
	if $(DOCKER_GO) -e GOOS=$(GOOS) -e GOARCH=$(GOARCH) $(GO_IMAGE) go build -o $$OUT .; \
	then \
		printf '\033[32mDone\033[0m in %sms, size %s\n' "$$(( $$($(NOW_MS)) - START ))" "$$(du -h $$OUT | cut -f1)"; \
	else \
		printf '\033[31mFailed\033[0m in %sms\n' "$$(( $$($(NOW_MS)) - START ))" >&2; \
		exit 1; \
	fi
