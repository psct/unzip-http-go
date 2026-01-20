.PHONY: build clean install test run-example

# Build the binary
build:
	go build -o unzip-http .

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o unzip-http-linux-amd64
	GOOS=darwin GOARCH=amd64 go build -o unzip-http-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o unzip-http-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -o unzip-http-windows-amd64.exe

# Install to $GOPATH/bin
install:
	go install .

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f unzip-http unzip-http-* *.exe

# Run example
run-example:
	go run . -l "https://www.learningcontainer.com/wp-content/uploads/2020/05/sample-zip-file.zip"
