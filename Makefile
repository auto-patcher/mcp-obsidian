.PHONY: build tidy test run

# Resolve/update dependencies and regenerate go.sum.
tidy:
	go mod tidy

# Build the server binary.
build: tidy
	go build -o bin/mcp-obsidian .

# Run tests.
test:
	go test ./...

# Build and run locally (OBSIDIAN_API_KEY must be set).
run: build
	./bin/mcp-obsidian
