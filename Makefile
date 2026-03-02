.PHONY: build run clean fmt lint test

build:
	go build -o bin/threedoors ./cmd/threedoors

run: build
	./bin/threedoors

clean:
	rm -rf bin/

fmt:
	gofumpt -w .

lint:
	golangci-lint run ./...

test:
	go test ./... -v
