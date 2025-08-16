BINARY_NAME=nodimus-memory

.PHONY: all test clean build release

all: build

test:
	go test ./... -v

clean:
	go clean
	rm -f $(BINARY_NAME)

build:
	go build -o $(BINARY_NAME) ./cmd/nodimus-memory

release:
	@echo "Creating new release..."
	@goreleaser release --rm-dist

readme:
	@echo "README.md has been created/updated. Please review it."
