build:
	mkdir -p ./target
	go build -o ./target/peacock ./cmd/peacock/

test:
	go test ./...
