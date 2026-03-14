.PHONY: build test test-cover lint clean

# Default target
all: build

# Build the binary
build:
	go build -o bin/oblivion ./cmd/oblivion

# Run all tests with the race detector in a single pass, collecting coverage
# for all packages (./...) and writing it to coverage.out so CI can ingest it.
# The threshold check uses the `total:` line from `go tool cover -func`, which
# is a statement-weighted average across all packages.  The main() stub (one
# statement calling os.Exit) is inherently untestable but contributes
# negligibly; the combined total still satisfies the 80% gate.
# The build fails when coverage drops below 80%.
test:
	go test -race -count=1 -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | tee /dev/stderr | \
		awk '/^total:/ { pct=$$3+0; if (pct < 80) { printf "FAIL: coverage %.1f%% < 80%%\n", pct; exit 1 } else { printf "OK:   coverage %.1f%% >= 80%%\n", pct } }'

# Open an HTML coverage report in the default browser (local development).
test-cover: test
	go tool cover -html=coverage.out

# Lint using golangci-lint if available, otherwise fall back to go vet.
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		go vet ./...; \
	fi

clean:
	rm -rf bin/ coverage.out
