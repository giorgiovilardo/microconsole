# Default task: List all available tasks
default:
    @just --list

# Run the application
run:
    go run .

# Run all tests
test:
    go test ./...

# Run tests with race condition detection
race:
    go test -race ./...

# Build the executable
build:
    mkdir -p dist
    go build -o dist/microconsole .

# Remove build artifacts
clean:
    rm -rf dist