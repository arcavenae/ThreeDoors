.PHONY: build run clean fmt lint test

build:
	go build -o bin/threedoors cmd/threedoors/main.go

run: build
	./bin/threedoors

clean:
	rm -rf bin/

fmt:
	gofumpt -l -w .

lint:
	golangci-lint run ./...

test:
	go test -v ./...
