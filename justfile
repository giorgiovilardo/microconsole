# Default task: List all available tasks
default:
    @just --list

# Run all tests
test:
    go test ./...

# Run tests with race condition detection
race:
    go test -race ./...

# Run tests in parallel with race detection
test-parallel:
    go test -v -race -parallel 8 ./...

# Run tests with coverage report
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

# Build the executable
build:
    mkdir -p dist
    go build -o dist/microconsole .

# Remove build artifacts
clean:
    rm -rf dist
