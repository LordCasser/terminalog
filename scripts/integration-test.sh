#!/bin/bash
# Terminalog Integration Test Script
# Creates a complete test environment in /tmp and runs full integration tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Determine project root directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Test environment directory
TEST_DIR="/tmp/terminalog-test-$(date +%Y%m%d-%H%M%S)"
SERVER_PID=""
SERVER_LOG="${TEST_DIR}/server.log"

echo -e "${BLUE}=== Terminalog Integration Test ===${NC}"
echo "Project root: ${PROJECT_ROOT}"
echo "Test directory: ${TEST_DIR}"

# Cleanup function
cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    if [ -n "$SERVER_PID" ] && kill -0 "$SERVER_PID" 2>/dev/null; then
        kill "$SERVER_PID" 2>/dev/null || true
        wait "$SERVER_PID" 2>/dev/null || true
    fi
    # Uncomment to preserve test environment for debugging
    # rm -rf "${TEST_DIR}"
    rm -rf "${TEST_DIR}"
    echo -e "${GREEN}Cleanup complete${NC}"
}

trap cleanup EXIT

# Step 1: Build the binary
echo -e "${BLUE}Step 1: Building binary...${NC}"
cd "${PROJECT_ROOT}"
make build
BINARY="${PROJECT_ROOT}/bin/terminalog"
if [ ! -f "$BINARY" ]; then
    echo -e "${RED}ERROR: Binary not found at ${BINARY}${NC}"
    exit 1
fi
echo -e "${GREEN}Binary built successfully${NC}"

# Step 2: Create test environment
echo -e "${BLUE}Step 2: Creating test environment...${NC}"
mkdir -p "${TEST_DIR}"
mkdir -p "${TEST_DIR}/articles"

# Create config.toml
cat > "${TEST_DIR}/config.toml" << 'EOF'
[server]
host = "localhost"
port = 18080

[blog]
content_dir = "./articles"

[logging]
level = "debug"
EOF

echo -e "${GREEN}Config created${NC}"

# Step 3: Create test markdown files
echo -e "${BLUE}Step 3: Creating test articles...${NC}"

# About Me
cat > "${TEST_DIR}/articles/_ABOUTME.md" << 'EOF'
# About Me

> Code is poetry, and I write in the language of machines.

## 👋 Hello

I'm a developer who loves clean code and terminal aesthetics.

## 🛠️ Tech Stack

- Go, TypeScript, Python
- React, Next.js, Tailwind CSS
- Git, Docker, Kubernetes

---

*Built with ❤️ using Terminalog*
EOF

# First article
cat > "${TEST_DIR}/articles/getting-started.md" << 'EOF'
# Getting Started with Terminalog

Welcome to Terminalog - a terminal-style blog system!

## Installation

```bash
# Clone the repository
git clone https://github.com/lordcasser/terminalog

# Build
make build

# Run
./bin/terminalog --content ./articles
```

## Features

- Markdown rendering
- Git history integration
- Version tracking
- Brutalist UI design

> "Simplicity is the ultimate sophistication." — Leonardo da Vinci

---

*EOF*
EOF

# Second article
cat > "${TEST_DIR}/articles/design-principles.md" << 'EOF'
# Design Principles

> Good design is as little design as possible.

## The Philosophy

1. **Clarity over cleverness**
2. **Explicit over implicit**
3. **Simple over easy**

## Brutalist Web Design

Brutalist design embraces:

- Raw, unpolished aesthetics
- Functional over decorative
- Truth to materials

```typescript
// Simple is better than complex
const greet = (name: string) => `Hello, ${name}!`;
```

---

*Design is not just what it looks like. Design is how it works.*
EOF

echo -e "${GREEN}Test articles created${NC}"

# Step 4: Initialize Git repository (server will do this automatically)
# But we'll add some commits to test version history
echo -e "${BLUE}Step 4: Setting up Git history...${NC}"
cd "${TEST_DIR}/articles"
git init
git config user.email "test@terminalog.dev"
git config user.name "Terminalog Test"
git add .
git commit -m "Initial articles"

# Add more commits for version history testing
cat >> "${TEST_DIR}/articles/getting-started.md" << 'EOF'

## Update 1

Added some updates for testing version history.

- Version tracking demo
- Git history integration test
EOF
git add .
git commit -m "Update getting-started article"

cat >> "${TEST_DIR}/articles/design-principles.md" << 'EOF'

## Practical Examples

Here's a practical example of the design principles:

### Single Responsibility

```go
// Good: Each function has one job
func ValidateUser(u *User) error { /* ... */ }
func SaveUser(u *User) error { /* ... */ }
func NotifyUser(u *User) error { /* ... */ }
```
EOF
git add .
git commit -m "Add practical examples"

echo -e "${GREEN}Git history setup complete (3 commits)${NC}"

# Step 5: Start server
echo -e "${BLUE}Step 5: Starting server...${NC}"
cd "${TEST_DIR}"
"${BINARY}" -config "${TEST_DIR}/config.toml" > "${SERVER_LOG}" 2>&1 &
SERVER_PID=$!
echo "Server PID: ${SERVER_PID}"

# Wait for server to start
sleep 3

# Check if server is running
if ! kill -0 "$SERVER_PID" 2>/dev/null; then
    echo -e "${RED}ERROR: Server failed to start${NC}"
    cat "${SERVER_LOG}"
    exit 1
fi

echo -e "${GREEN}Server started${NC}"

# Step 6: Test API endpoints
echo -e "${BLUE}Step 6: Testing API endpoints...${NC}"

BASE_URL="http://localhost:18080"

# Test 1: Health check
echo -e "${YELLOW}Test 1: Health check${NC}"
HEALTH=$(curl -s "${BASE_URL}/healthz")
if [ -z "$HEALTH" ]; then
    echo -e "${RED}FAIL: Health check failed${NC}"
    exit 1
fi
echo -e "${GREEN}PASS: Health check${NC}"

# Test 2: Articles list
echo -e "${YELLOW}Test 2: Articles list${NC}"
ARTICLES=$(curl -s "${BASE_URL}/api/articles")
if ! echo "$ARTICLES" | grep -q "getting-started"; then
    echo -e "${RED}FAIL: Articles list missing getting-started${NC}"
    echo "$ARTICLES"
    exit 1
fi
if ! echo "$ARTICLES" | grep -q "design-principles"; then
    echo -e "${RED}FAIL: Articles list missing design-principles${NC}"
    exit 1
fi
# Verify _ABOUTME.md is excluded
if echo "$ARTICLES" | grep -q "_ABOUTME"; then
    echo -e "${RED}FAIL: _ABOUTME.md should be excluded from articles list${NC}"
    exit 1
fi
echo -e "${GREEN}PASS: Articles list (2 articles, _ABOUTME excluded)${NC}"

# Test 3: Article content
echo -e "${YELLOW}Test 3: Article content${NC}"
CONTENT=$(curl -s "${BASE_URL}/api/articles/getting-started.md")
if ! echo "$CONTENT" | grep -q "Getting Started"; then
    echo -e "${RED}FAIL: Article content missing title${NC}"
    exit 1
fi
echo -e "${GREEN}PASS: Article content${NC}"

# Test 4: About Me
echo -e "${YELLOW}Test 4: About Me API${NC}"
ABOUTME=$(curl -s "${BASE_URL}/api/aboutme")
if ! echo "$ABOUTME" | grep -q "About Me"; then
    echo -e "${RED}FAIL: About Me content missing${NC}"
    exit 1
fi
echo -e "${GREEN}PASS: About Me API${NC}"

# Test 5: Version API
echo -e "${YELLOW}Test 5: Version API${NC}"
VERSION=$(curl -s "${BASE_URL}/api/articles/getting-started.md/version")
if ! echo "$VERSION" | grep -q "currentVersion"; then
    echo -e "${RED}FAIL: Version API failed${NC}"
    echo "$VERSION"
    exit 1
fi
if ! echo "$VERSION" | grep -q "v1"; then
    echo -e "${RED}FAIL: Version should start with v1${NC}"
    exit 1
fi
echo -e "${GREEN}PASS: Version API${NC}"

# Test 6: Frontend homepage
echo -e "${YELLOW}Test 6: Frontend homepage${NC}"
HOMEPAGE=$(curl -s "${BASE_URL}/")
if ! echo "$HOMEPAGE" | grep -q "Terminalog"; then
    echo -e "${RED}FAIL: Frontend missing title${NC}"
    exit 1
fi
echo -e "${GREEN}PASS: Frontend homepage${NC}"

# Test 7: Frontend aboutme page
echo -e "${YELLOW}Test 7: Frontend aboutme page${NC}"
ABOUTME_PAGE=$(curl -s "${BASE_URL}/aboutme/")
if ! echo "$ABOUTME_PAGE" | grep -q "<!DOCTYPE html>"; then
    echo -e "${RED}FAIL: Aboutme page not valid HTML${NC}"
    echo "Response content:"
    echo "$ABOUTME_PAGE" | head -20
    echo "---"
    echo "Server log:"
    head -30 "${SERVER_LOG}"
    exit 1
fi
echo -e "${GREEN}PASS: Frontend aboutme page${NC}"

# Test 8: Frontend article page
echo -e "${YELLOW}Test 8: Frontend article page${NC}"
ARTICLE_PAGE=$(curl -s "${BASE_URL}/article/?path=getting-started.md")
if ! echo "$ARTICLE_PAGE" | grep -q "<!DOCTYPE html>"; then
    echo -e "${RED}FAIL: Article page not valid HTML${NC}"
    exit 1
fi
echo -e "${GREEN}PASS: Frontend article page${NC}"

# Step 7: Summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo -e "${GREEN}All 8 tests passed!${NC}"
echo ""
echo "Test environment preserved at: ${TEST_DIR}"
echo "To manually test:"
echo "  - Server URL: ${BASE_URL}"
echo "  - Server log: ${SERVER_LOG}"
echo ""
echo -e "${GREEN}Integration test complete!${NC}"