VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X github.com/dubeyKartikay/peacock/internal/cli.Version=$(VERSION)"

build:
	mkdir -p ./target
	go build $(LDFLAGS) -o ./target/peacock ./cmd/peacock/

test:
	go test ./...
