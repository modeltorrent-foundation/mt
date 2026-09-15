# Model Torrent — developer entrypoints.

.PHONY: test test-go test-py fixtures swarm-smoke build clean tidy

## test: run Go and Python tests.
test: test-go test-py

## test-go: run all Go package tests.
test-go:
	go test ./...

## test-py: run the Python shim tests. Skipped (not failed) if pytest is absent.
test-py:
	@if python3 -m pytest --version >/dev/null 2>&1; then \
		python3 -m pytest shim/tests -q; \
	else \
		echo "pytest not installed; skipping Python shim tests"; \
	fi

## fixtures: (re)generate tiny KB stand-in models + dev catalog index.
fixtures:
	go run ./scripts/gen_fixtures.go
	go run ./scripts/gen_catalog.go
	go run ./scripts/gen_web_magnets.go

## swarm-smoke: focused localhost seed/fetch test.
swarm-smoke:
	go test ./internal/torrentsvc -run TestSwarmSmoke -v

## build: compile the mt CLI to ./bin/mt.
build:
	go build -o bin/mt ./cmd/mt

## tidy: keep go.mod honest.
tidy:
	go mod tidy

## clean: remove build output and generated fixtures.
clean:
	rm -rf bin testdata/fixtures
