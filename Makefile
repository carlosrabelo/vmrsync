MAKEFLAGS += --no-print-directory

.PHONY: build test lint fmt clean install uninstall help

help:
	@echo "Available targets:"
	@echo "  build     - Compile the binary to bin/"
	@echo "  test      - Run Go unit tests"
	@echo "  lint      - Run go vet"
	@echo "  fmt       - Format code with gofmt"
	@echo "  clean     - Remove build artifacts"
	@echo "  install   - Install binary to ~/.local/bin"
	@echo "  uninstall - Remove binary from ~/.local/bin"

build:
	@./.make/build.sh

test:
	@./.make/test.sh

lint:
	@go vet ./...

fmt:
	@go fmt ./...

clean:
	@go clean
	@rm -rf bin/

install: build
	@./.make/install.sh

uninstall:
	@./.make/uninstall.sh
