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
FRONTEND_DIR=frontend
STATIC_DIR=pkg/embed/static

# Version info
VERSION?=dev
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE)"

.PHONY: all build clean run test help frontend backend

all: build

## build: Build the complete application (frontend + backend)
build: backend

## backend: Build the Go backend binary
backend:
	@echo "Building backend..."
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) cmd/terminalog/main.go
	@echo "Backend built: $(BIN_DIR)/$(BINARY_NAME)"

## frontend: Build the Next.js frontend (copy to embed directory)
frontend:
	@echo "Building frontend..."
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && npm run build && \
		cp -r out/* ../$(STATIC_DIR)/; \
		echo "Frontend built and copied to $(STATIC_DIR)"; \
	else \
		echo "Frontend directory not found. Skipping."; \
	fi

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
	rm -rf $(FRONTEND_DIR)/.next
	rm -rf $(FRONTEND_DIR)/out

## tidy: Tidy Go modules
tidy:
	@echo "Tidy Go modules..."
	$(GOMOD) tidy

## deps: Install dependencies
deps:
	@echo "Installing Go dependencies..."
	$(GOMOD) download
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && npm install; \
	fi

## release: Build release binaries for multiple platforms
release:
	@echo "Building release binaries..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_LINUX) cmd/terminalog/main.go
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_WINDOWS) cmd/terminalog/main.go
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_DARWIN) cmd/terminalog/main.go
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