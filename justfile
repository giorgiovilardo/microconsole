# Default task: List all available tasks
default:
    @just --list

# Formats and lints (basic)
lint:
    gofmt -l -w .
    go vet ./...

# Run all tests fast
test:
    go test ./...

# Run all tests with race condition verifier
test-race:
    go test -race ./...

# Run all tests with coverage
test-coverage:
    go test -coverprofile coverage.out ./...

# Run tests with coverage report and enforce threshold
test-coverage-check threshold="77":
    just test-coverage
    @COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//'); \
    if [ $(echo "$COVERAGE < {{threshold}}" | bc -l) -eq 1 ]; then \
        echo "Coverage $COVERAGE% is below threshold {{threshold}}%"; \
        exit 1; \
    fi; \
    echo "Coverage $COVERAGE% meets threshold {{threshold}}%"
