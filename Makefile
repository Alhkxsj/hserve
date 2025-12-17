# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get
GOVET=$(GOCMD) vet
GOFMT=gofmt
BINARY_NAME=hserve
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the project
build: 
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/hserve

# Install the binary to system
install: build
	cp $(BINARY_NAME) $(HOME)/go/bin/ || cp $(BINARY_NAME) /usr/local/bin/ || echo "Please copy $(BINARY_NAME) to a directory in your PATH"

# Run tests
test: 
	$(GOTEST) -v ./...

# Run go vet
vet:
	$(GOVET) ./...

# Format code
fmt:
	$(GOFMT) -s -w ./

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Run go mod tidy
tidy:
	$(GOMOD) tidy

# Build for multiple architectures
multiarch:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_NAME)-linux-amd64 -v ./cmd/hserve
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o dist/$(BINARY_NAME)-linux-arm64 -v ./cmd/hserve
	GOOS=linux GOARCH=arm $(GOBUILD) -o dist/$(BINARY_NAME)-linux-arm -v ./cmd/hserve
	GOOS=android GOARCH=arm64 $(GOBUILD) -o dist/$(BINARY_NAME)-android-arm64 -v ./cmd/hserve
	GOOS=android GOARCH=arm $(GOBUILD) -o dist/$(BINARY_NAME)-android-arm -v ./cmd/hserve
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_NAME)-darwin-amd64 -v ./cmd/hserve
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o dist/$(BINARY_NAME)-darwin-arm64 -v ./cmd/hserve
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o dist/$(BINARY_NAME)-windows-amd64.exe -v ./cmd/hserve

# Build deb package
deb:
	@echo "Building deb package..."
	@mkdir -p dist
	./scripts/build-deb.sh

# Install deb package
install-deb: deb
	sudo dpkg -i dist/*.deb

# Run all checks
check: vet test

# Generate certificates (for testing)
gen-cert:
	./$(BINARY_NAME) gen-cert

# Run server (for testing)
serve:
	./$(BINARY_NAME) serve

.PHONY: build install test vet fmt clean multiarch deb install-deb check gen-cert serve tidy