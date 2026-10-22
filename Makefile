.DEFAULT_GOAL := help
.SILENT:
.PHONY: help build build-all

# Step headline, example: $(HEADLINE) 'Example headline'
# Uses printf, because escape handling in echo differs per shell and may not expand \033
HEADLINE = printf '\n\033[7m \# \033[0m \033[1m%s\033[0m\n'

# Go version (read from go.mod)
GO_VERSION := $(shell awk '/^go /{print $$2}' go.mod)
GO_IMAGE ?= golang:$(GO_VERSION)

# Target platform, defaults to Apple Silicon Mac. Example: make build GOOS=linux GOARCH=amd64
GOOS ?= darwin
GOARCH ?= arm64
BINARY ?= walker
PLATFORMS := darwin/arm64 darwin/amd64 linux/amd64 linux/arm64 windows/amd64

EXT = $(if $(filter windows,$(GOOS)),.exe)
OUT = dist/$(BINARY)-$(GOOS)-$(GOARCH)$(EXT)

help: ## Show available commands
	grep -hE '^[a-zA-Z][a-zA-Z-]*:.*## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*## "}; {printf "  \033[32m%-24s\033[0m %s\n", $$1, $$2}'

build: ## Build binary for one platform (GOOS, GOARCH), default darwin/arm64
	$(HEADLINE) 'Build $(GOOS)/$(GOARCH)'
	mkdir -p dist
	docker run --rm \
		-v "$(CURDIR)":/app -w /app \
		-v go-mod:/go/pkg/mod \
		-v go-build:/root/.cache/go-build \
		-e GOOS=$(GOOS) -e GOARCH=$(GOARCH) \
		$(GO_IMAGE) go build -o $(OUT) .
	printf 'Result: \033[36m%s\033[0m\n' '$(OUT)'

build-all: ## Build binaries for all supported platforms
	for p in $(PLATFORMS); do \
		$(MAKE) build GOOS=$${p%/*} GOARCH=$${p#*/} || exit 1; \
	done
