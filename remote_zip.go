package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RemoteZipFile represents a ZIP file accessed via HTTP
type RemoteZipFile struct {
	URL        string
	httpClient *http.Client
	size       int64
	files      []*zip.File
	reader     *zip.Reader
}

// NewRemoteZipFile creates a new RemoteZipFile instance
func NewRemoteZipFile(url string) (*RemoteZipFile, error) {
	// Create HTTP client with connection pooling and keep-alive
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
		DisableCompression:  true, // We handle compression ourselves
	}
	
	rzf := &RemoteZipFile{
		URL: url,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}

	// Get the file size
	resp, err := rzf.httpClient.Head(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Check if server supports range requests
	if resp.Header.Get("Accept-Ranges") != "bytes" {
		return nil, fmt.Errorf("server does not support range requests")
	}

	rzf.size = resp.ContentLength
	if rzf.size <= 0 {
		return nil, fmt.Errorf("could not determine file size")
	}

	// Read the central directory
	if err := rzf.readCentralDirectory(); err != nil {
		return nil, fmt.Errorf("failed to read central directory: %w", err)
	}

	return rzf, nil
}

// Close closes the HTTP client and cleans up resources
func (rzf *RemoteZipFile) Close() {
	if rzf.httpClient != nil && rzf.httpClient.Transport != nil {
		if transport, ok := rzf.httpClient.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
}

// getRange retrieves a specific byte range from the remote file
func (rzf *RemoteZipFile) getRange(start, end int64) ([]byte, error) {
	req, err := http.NewRequest("GET", rzf.URL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end-1))

	resp, err := rzf.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// readCentralDirectory reads the ZIP central directory from the end of the file
func (rzf *RemoteZipFile) readCentralDirectory() error {
	// ZIP files have the End of Central Directory (EOCD) record at the end
	// We'll read the last 64KB to be safe (accounts for comments)
	searchSize := int64(65536)
	if searchSize > rzf.size {
		searchSize = rzf.size
	}

	// Read the end of the file
	endData, err := rzf.getRange(rzf.size-searchSize, rzf.size)
	if err != nil {
		return err
	}

	// Find the End of Central Directory signature (0x06054b50)
	eocdSignature := []byte{0x50, 0x4b, 0x05, 0x06}
	eocdPos := -1
	for i := len(endData) - 22; i >= 0; i-- {
		if bytes.Equal(endData[i:i+4], eocdSignature) {
			eocdPos = i
			break
		}
	}

	if eocdPos < 0 {
		return fmt.Errorf("could not find End of Central Directory record")
	}

	// Parse EOCD to find central directory location
	eocd := endData[eocdPos:]
	if len(eocd) < 22 {
		return fmt.Errorf("EOCD record too short")
	}

	// Create a custom ReaderAt that can read from remote ranges
	readerAt := &remoteReaderAt{rzf: rzf}

	// Parse the ZIP structure
	zipReader, err := zip.NewReader(readerAt, rzf.size)
	if err != nil {
		return err
	}

	rzf.reader = zipReader
	rzf.files = zipReader.File

	return nil
}

// List returns a list of file names in the ZIP archive
func (rzf *RemoteZipFile) List() []string {
	names := make([]string, len(rzf.files))
	for i, f := range rzf.files {
		names[i] = f.Name
	}
	return names
}

// Files returns the list of files in the ZIP archive
func (rzf *RemoteZipFile) Files() []*zip.File {
	return rzf.files
}

// Open opens a file from the ZIP archive and returns a ReadCloser
func (rzf *RemoteZipFile) Open(name string) (io.ReadCloser, error) {
	for _, f := range rzf.files {
		if f.Name == name {
			return f.Open()
		}
	}

	return nil, fmt.Errorf("file not found: %s", name)
}

// Extract extracts a file to the specified output path
func (rzf *RemoteZipFile) Extract(name string) ([]byte, error) {
	rc, err := rzf.Open(name)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return io.ReadAll(rc)
}

// remoteReaderAt implements io.ReaderAt for remote ZIP file access
type remoteReaderAt struct {
	rzf *RemoteZipFile
}

func (r *remoteReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	data, err := r.rzf.getRange(off, off+int64(len(p)))
	if err != nil {
		return 0, err
	}
	copy(p, data)
	return len(data), nil
}
