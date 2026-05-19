BINARY := deep-search-mcp
PKG    := ./cmd/deep-search-mcp

.PHONY: build test vet tidy run clean

build:
	go build -o bin/$(BINARY) $(PKG)

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

run: build
	./bin/$(BINARY)

clean:
	rm -rf bin
