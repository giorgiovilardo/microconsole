# Default task: List all available tasks
default:
    @just --list

# Run all tests
test:
    go test -race ./...

# Run tests with coverage report
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out
