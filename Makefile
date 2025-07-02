.PHONY: build run run-stdio run-sse run-streamable clean test fmt lint

# Binary name.
BINARY_NAME=d2mcp

# Build the binary.
build:
	go build -o $(BINARY_NAME) ./cmd

# Run the server.
run: build
	./$(BINARY_NAME)

# Run with STDIO transport.
run-stdio: build
	./$(BINARY_NAME) -transport=stdio

# Run with SSE transport.
run-sse: build
	./$(BINARY_NAME) -transport=sse

# Run with Streamable HTTP transport.
run-streamable: build
	./$(BINARY_NAME) -transport=streamable

# Clean build artifacts.
clean:
	rm -f $(BINARY_NAME)

# Run tests.
test:
	go test -v ./...

# Format code.
fmt:
	go fmt ./...

# Run linter.
lint:
	golangci-lint run

# Install dependencies.
deps:
	go mod download
	go mod tidy

# Build for multiple platforms.
build-all:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 ./cmd
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 ./cmd
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 ./cmd
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe ./cmd
