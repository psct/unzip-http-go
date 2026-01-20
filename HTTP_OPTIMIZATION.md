# HTTP Connection Optimization

## How it works

The Go implementation now optimizes HTTP connections through:

### 1. Connection Pooling
```go
MaxIdleConns:        10,
MaxIdleConnsPerHost: 10,
```
- Maintains up to 10 idle connections per host
- Reuses connections instead of creating new ones for each request

### 2. HTTP Keep-Alive
```go
DisableKeepAlives: false,
IdleConnTimeout:   90 * time.Second,
```
- Keeps TCP connections open for 90 seconds
- Server responds with `Connection: keep-alive` header
- Same TCP connection is reused for multiple HTTP requests

### 3. Compression Handling
```go
DisableCompression: true,
```
- We disable automatic compression because ZIP files are already compressed
- Prevents double-compression overhead

## Expected Behavior

For a typical use case:
1. **HEAD request** - Establish connection, get file size
2. **GET request (range)** - Reuse connection, download EOCD
3. **Multiple GET requests** - All reuse the same TCP connection for extracting files

This means:
- ✅ **Single TCP connection** for all operations (if server supports Keep-Alive)
- ✅ **Faster subsequent requests** (no TCP handshake overhead)
- ✅ **Lower server load** (fewer connections)

## Verifying Connection Reuse

You can verify connection reuse with packet capture:
```bash
# On Windows with Wireshark or on Linux:
tcpdump -i any host example.com

# Look for TCP handshakes (SYN/SYN-ACK/ACK)
# You should see only ONE handshake for the entire session
```

## Server Requirements

The remote HTTP server must support:
- `Accept-Ranges: bytes` (required for range requests)
- `Connection: keep-alive` (recommended, most servers support this)
- HTTP/1.1 persistent connections (standard)

Most modern web servers (nginx, Apache, IIS, S3, etc.) support all of these by default.
