# Terminalog - Terminal-style Blog System
# Makefile for build and development

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
BINARY_NAME=terminalog
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_LINUX=$(BINARY_NAME)_linux
BINARY_WINDOWS=$(BINARY_NAME)_windows.exe
BINARY_DARWIN=$(BINARY_NAME)_darwin

# Directories
BIN_DIR=bin
WEB_DIR=frontend
STATIC_DIR=pkg/embed/static

# Version info
VERSION?=dev
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE)"

.PHONY: all build clean run test help frontend backend web-embed

all: build

## build: Build the complete application (frontend + backend)
build: web-embed backend

## web-embed: Build frontend and copy to embed directory
web-embed: frontend
	@echo "Copying frontend build to embed directory..."
	@rm -rf $(STATIC_DIR)/*
	@mkdir -p $(STATIC_DIR)
	@if [ -d "$(WEB_DIR)/out" ]; then \
		cp -r $(WEB_DIR)/out/* $(STATIC_DIR)/; \
		find $(STATIC_DIR) -type d -empty -exec touch {}/.gitkeep \; ; \
		echo "Frontend copied to $(STATIC_DIR)"; \
	else \
		echo "Warning: Frontend build output not found at $(WEB_DIR)/out"; \
		touch $(STATIC_DIR)/index.html; \
		echo '<html><body><h1>Frontend not built</h1></body></html>' > $(STATIC_DIR)/index.html; \
	fi

## frontend: Build the Next.js frontend (static export)
frontend:
	@echo "Building frontend..."
	@if [ -d "$(WEB_DIR)" ]; then \
		cd $(WEB_DIR) && npm install && npm run build; \
		echo "Frontend build complete"; \
	else \
		echo "Error: Frontend directory not found at $(WEB_DIR)"; \
		exit 1; \
	fi

## backend: Build the Go backend binary (with embedded frontend)
backend:
	@echo "Building backend with embedded frontend..."
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) cmd/terminalog/main.go
	@echo "Backend built: $(BIN_DIR)/$(BINARY_NAME)"

## run: Run the application
run:
	@echo "Running terminalog..."
	./$(BIN_DIR)/$(BINARY_NAME) --log debug

## dev: Run in development mode
dev:
	@echo "Running in development mode..."
	$(GOCMD) run cmd/terminalog/main.go --log debug

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BIN_DIR)
	rm -rf $(WEB_DIR)/.next
	rm -rf $(WEB_DIR)/out
	rm -rf $(STATIC_DIR)

## tidy: Tidy Go modules
tidy:
	@echo "Tidy Go modules..."
	$(GOMOD) tidy

## deps: Install dependencies
deps:
	@echo "Installing Go dependencies..."
	$(GOMOD) download
	@if [ -d "$(WEB_DIR)" ]; then \
		cd $(WEB_DIR) && npm install; \
	fi

## release: Build release binaries for multiple platforms (using goreleaser)
release:
	@echo "Building release with goreleaser..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean; \
	else \
		echo "goreleaser not found. Install it first."; \
		echo "Alternative: use 'make release-manual'"; \
	fi

## release-manual: Build release binaries manually (without goreleaser)
release-manual: web-embed
	@echo "Building release binaries manually..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_LINUX) cmd/terminalog/main.go
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_WINDOWS) cmd/terminalog/main.go
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_DARWIN) cmd/terminalog/main.go
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME)_darwin_arm64 cmd/terminalog/main.go
	@echo "Release binaries built in $(BIN_DIR)/"

## install: Install the binary to system
install:
	@echo "Installing..."
	cp $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installed to /usr/local/bin/$(BINARY_NAME)"

## help: Show this help message
help:
	@echo "Terminalog - Terminal-style Blog System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## /  /p' $(MAKEFILE_LIST)

# Default target
.DEFAULT_GOAL := build