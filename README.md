# unzip-http-go

Go implementation of [unzip-http](https://github.com/saulpw/unzip-http) - Extract individual files from .zip files over HTTP without downloading the entire archive.

## Features

- Extract single files from remote ZIP archives using HTTP range requests
- List contents of remote ZIP files without downloading
- Support for wildcard patterns
- Write to stdout or extract to disk
- Recreate folder structure or flatten to current directory

## Requirements

- Go 1.21 or higher
- HTTP server must send `Accept-Ranges: bytes` and `Content-Length` headers (most do)

## Installation

```bash
go build -o unzip-http
```

Or install directly:

```bash
go install github.com/unzip-http-go@latest
```

## Usage

### Command Line

```bash
# List files in remote ZIP
unzip-http -l https://example.com/archive.zip

# Extract specific file
unzip-http https://example.com/archive.zip README.txt

# Extract multiple files
unzip-http https://example.com/archive.zip file1.txt file2.txt

# Extract with wildcard pattern
unzip-http https://example.com/archive.zip "*.txt"

# Recreate folder structure
unzip-http -f https://example.com/archive.zip docs/manual.pdf

# Write to stdout
unzip-http -o https://example.com/archive.zip data.json
```

### As a Library

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    // Open remote ZIP file
    rzf, err := NewRemoteZipFile("https://example.com/archive.zip")
    if err != nil {
        log.Fatal(err)
    }

    // List files
    files := rzf.List()
    for _, name := range files {
        fmt.Println(name)
    }

    // Extract a specific file
    data, err := rzf.Extract("README.txt")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("File content: %s\n", data)

    // Open file for reading
    rc, err := rzf.Open("data.json")
    if err != nil {
        log.Fatal(err)
    }
    defer rc.Close()
    
    // Read from rc...
}
```

## How It Works

The tool uses HTTP range requests to:

1. First, make a HEAD request to get the file size and verify range support
2. Download only the last ~64KB of the ZIP file to read the Central Directory
3. Parse the Central Directory to get file locations
4. When extracting, download only the specific bytes for requested files

This means that for a 1GB ZIP file, you might only download a few KB to list contents, or a few MB to extract a single small file.

## Options

- `-l` - List files in remote .zip file (default if no filenames given)
- `-f` - Recreate folder structure from .zip file when extracting (instead of extracting files to the current directory)
- `-o` - Write files to stdout (if multiple files, concatenate them in zipfile order)

## Comparison with Python Version

This Go implementation provides the same core functionality as the original Python `unzip-http`:

- ✅ HTTP range request support
- ✅ List remote ZIP contents
- ✅ Extract individual files
- ✅ Wildcard pattern matching
- ✅ Folder structure recreation
- ✅ Stdout output
- ✅ Library usage

The Go version offers:
- Better performance (compiled binary)
- No runtime dependencies
- Cross-platform single binary
- Type safety

## Examples

```bash
# List files in a remote ZIP
./unzip-http -l "https://github.com/example/repo/archive/refs/heads/main.zip"

# Extract README from GitHub archive
./unzip-http "https://github.com/example/repo/archive/refs/heads/main.zip" "*/README.md"

# Extract all text files and pipe to grep
./unzip-http -o "https://example.com/data.zip" "*.txt" | grep "searchterm"
```

## License

MIT License

## Credits

Original Python implementation by [Saul Pwanson](https://saul.pw)
Go port created using Claude
