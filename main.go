package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Command-line flags
	listFiles := flag.Bool("l", false, "List files in remote .zip file")
	recreateStructure := flag.Bool("f", false, "Recreate folder structure from .zip file when extracting")
	writeStdout := flag.Bool("o", false, "Write files to stdout")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: unzip-http [-l] [-f] [-o] <url> [filenames...]\n")
		fmt.Fprintf(os.Stderr, "\nExtract individual files from .zip files over http without downloading the entire archive.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -l    List files in remote .zip file (default if no filenames given)\n")
		fmt.Fprintf(os.Stderr, "  -f    Recreate folder structure from .zip file when extracting\n")
		fmt.Fprintf(os.Stderr, "  -o    Write files to stdout\n")
		os.Exit(1)
	}

	url := args[0]
	filenames := args[1:]

	// Create RemoteZipFile
	rzf, err := NewRemoteZipFile(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer rzf.Close()

	// If no filenames provided or -l flag is set, list files
	if *listFiles || len(filenames) == 0 {
		listZipContents(rzf)
		return
	}

	// Extract requested files
	for _, pattern := range filenames {
		if err := extractFiles(rzf, pattern, *recreateStructure, *writeStdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error extracting %s: %v\n", pattern, err)
		}
	}
}

func listZipContents(rzf *RemoteZipFile) {
	fmt.Printf("%-10s  %-19s  %s\n", "Length", "DateTime", "Name")
	fmt.Println(strings.Repeat("-", 60))

	for _, f := range rzf.Files() {
		fmt.Printf("%-10d  %s  %s\n",
			f.UncompressedSize64,
			f.Modified.Format("2006-01-02 15:04:05"),
			f.Name)
	}
}

func extractFiles(rzf *RemoteZipFile, pattern string, recreateStructure, writeStdout bool) error {
	matched := false

	for _, f := range rzf.Files() {
		// Normalize the file name from the ZIP (always uses forward slashes)
		normalizedName := filepath.FromSlash(f.Name)
		
		// Simple pattern matching (supports * wildcard)
		if matchPattern(pattern, f.Name) || matchPattern(pattern, normalizedName) {
			matched = true

			if f.FileInfo().IsDir() {
				continue
			}

			if writeStdout {
				// Write to stdout
				data, err := rzf.Extract(f.Name)
				if err != nil {
					return fmt.Errorf("failed to extract %s: %w", f.Name, err)
				}
				os.Stdout.Write(data)
			} else {
				// Write to file
				outputPath := normalizedName
				if !recreateStructure {
					outputPath = filepath.Base(normalizedName)
				}

				// Create directory structure if needed
				dir := filepath.Dir(outputPath)
				if dir != "." && dir != "" {
					if err := os.MkdirAll(dir, 0755); err != nil {
						return fmt.Errorf("failed to create directory %s: %w", dir, err)
					}
				}

				fmt.Fprintf(os.Stderr, "Extracting %s...\n", f.Name)

				data, err := rzf.Extract(f.Name)
				if err != nil {
					return fmt.Errorf("failed to extract %s: %w", f.Name, err)
				}

				if err := os.WriteFile(outputPath, data, 0644); err != nil {
					return fmt.Errorf("failed to write %s: %w", outputPath, err)
				}
			}
		}
	}

	if !matched {
		return fmt.Errorf("no files matched pattern: %s", pattern)
	}

	return nil
}

// Simple pattern matching with * wildcard support
func matchPattern(pattern, name string) bool {
	// Normalize both pattern and name to use forward slashes for comparison
	pattern = filepath.ToSlash(pattern)
	name = filepath.ToSlash(name)
	
	if pattern == name {
		return true
	}

	// Handle wildcard patterns
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			// Simple prefix*suffix matching
			return strings.HasPrefix(name, parts[0]) && strings.HasSuffix(name, parts[1])
		}
	}

	return false
}
