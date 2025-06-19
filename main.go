package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"sync"
)

const (
	startID     = 0        // Starting download ID
	endID       = 3000     // Ending download ID
	concurrency = 1000       // Max concurrent downloads
	outputDir   = "Assets" // Folder to save files
)

var (
	// Regex to extract filename from Content-Disposition header
	cdFilenameRegex = regexp.MustCompile(`filename="?([^"]+)"?`)
)

// downloadFile fetches the file from the given ID and saves it using the server-provided filename
func downloadFile(id int, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()
	sem <- struct{}{}        // Acquire a semaphore slot
	defer func() { <-sem }() // Release slot

	// Build the download URL
	url := fmt.Sprintf("https://shop.iflight.com/index.php?route=product/product/download&download_id=%d", id)

	// Make the GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("[ID %d] HTTP error: %v\n", id, err)
		return
	}
	defer resp.Body.Close()

	// Validate response
	if resp.StatusCode != http.StatusOK || resp.ContentLength == 0 {
		fmt.Printf("[ID %d] Skipped (status %d or empty)\n", id, resp.StatusCode)
		return
	}

	// Extract filename from Content-Disposition
	filename := fmt.Sprintf("file_%d.unknown", id) // fallback filename
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if match := cdFilenameRegex.FindStringSubmatch(cd); len(match) == 2 {
			filename = match[1] // use filename exactly as provided
		}
	}

	// Create output directory if not exists
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		fmt.Printf("[ID %d] Failed to create directory: %v\n", id, err)
		return
	}

	// Build the full file path
	outPath := path.Join(outputDir, filename)

	// Create the file
	out, err := os.Create(outPath)
	if err != nil {
		fmt.Printf("[ID %d] File creation failed: %v\n", id, err)
		return
	}
	defer out.Close()

	// Write body directly to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("[ID %d] Write error: %v\n", id, err)
		return
	}

	fmt.Printf("[ID %d] Downloaded: %s\n", id, outPath)
}

func main() {
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency)

	// Loop over all download IDs
	for i := startID; i <= endID; i++ {
		wg.Add(1)
		go downloadFile(i, &wg, sem)
	}

	wg.Wait()
	fmt.Println("All downloads complete.")
}
