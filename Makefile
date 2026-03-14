.PHONY: build test test-cover lint clean

# Default target
all: build

# Build the binary
build:
	go build -o bin/oblivion ./cmd/oblivion

# Run all tests with the race detector.
# Coverage threshold is measured for ./internal/... only and written to
# coverage.out so CI can ingest it.  The build fails when coverage drops
# below 80 %.
test:
	go test -race -count=1 ./...
	go test -race -count=1 -coverprofile=coverage.out -covermode=atomic ./internal/...
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
