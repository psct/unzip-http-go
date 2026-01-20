package main

import (
	"fmt"
	"log"
)

// Example demonstrates how to use the RemoteZipFile library
func ExampleUsage() {
	// Example 1: List files in a remote ZIP
	fmt.Println("Example 1: List files")
	rzf, err := NewRemoteZipFile("https://example.com/archive.zip")
	if err != nil {
		log.Fatal(err)
	}

	files := rzf.List()
	fmt.Printf("Found %d files:\n", len(files))
	for _, name := range files {
		fmt.Printf("  - %s\n", name)
	}

	// Example 2: Get detailed file information
	fmt.Println("\nExample 2: File details")
	for _, f := range rzf.Files() {
		fmt.Printf("%s: %d bytes (compressed: %d bytes)\n",
			f.Name, f.UncompressedSize64, f.CompressedSize64)
	}

	// Example 3: Extract a specific file to memory
	fmt.Println("\nExample 3: Extract file")
	data, err := rzf.Extract("README.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Extracted %d bytes\n", len(data))
	fmt.Printf("Content preview: %s\n", data[:min(100, len(data))])

	// Example 4: Stream a file
	fmt.Println("\nExample 4: Stream file")
	rc, err := rzf.Open("data.json")
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()

	// Read from rc as needed...
	fmt.Println("File opened successfully")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
