.PHONY: build run test clean tidy

BINARY=dockercode
MAIN=.

build:
	go build -o $(BINARY) $(MAIN)

run:
	go run $(MAIN)

test:
	go test ./... -v -race

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)

lint:
	golangci-lint run ./...
