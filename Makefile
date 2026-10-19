VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X github.com/dubeyKartikay/peacock/internal/cli.Version=$(VERSION)"

.PHONY: build test test-ui release-check release-snapshot stress

build:
	mkdir -p ./target
	go build $(LDFLAGS) -o ./target/peacock ./cmd/peacock/

test:
	go test ./...

test-ui:
	python3 scripts/test-ui-vhs.py run

release-check:
	go mod verify
	go test -race ./... ./testdata/stress-test/...
	go vet ./... ./testdata/stress-test/...
	goreleaser check

release-snapshot: release-check
	goreleaser release --snapshot --clean
	python3 scripts/homebrew-formula.py

stress: build
	go run ./testdata/stress-test/main.go | ./target/peacock
