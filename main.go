package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
)

func main() {
	// Define the working directory where files will be saved.
	workingDirectory := "Assets/"

	// Check if the working directory exists; if not, create it with permission 0700.
	if !directoryExists(workingDirectory) {
		createDirectory(workingDirectory, 0700)
	}

	// Loop from 0 up to the maximum download ID (loopStopCounter).
	// You can increase this if needed.
	loopStopCounter := 1000

	// Iterate through the download IDs.
	for id := 0; id <= loopStopCounter; id++ {
		// Construct the download URL dynamically.
		url := fmt.Sprintf("https://shop.iflight.com/index.php?route=product/product/download&download_id=%d", id)
		// Download the file for this URL.
		downloadFiles(url, workingDirectory)
	}
}

// downloadFiles sends an HTTP GET request to the given URL and saves the response body to disk.
func downloadFiles(givenURL string, workingDirectory string) {
	// Send a GET request to the URL.
	response, err := http.Get(givenURL)
	if err != nil {
		log.Printf("Error making request to %s: %v\n", givenURL, err)
		return
	}
	defer response.Body.Close() // Ensure the response body is closed later.

	// If the HTTP response code is not 200 OK, skip this file.
	if response.StatusCode != http.StatusOK || response.ContentLength == 0 {
		log.Printf("Skipping: %s - StatusCode: %d\n", givenURL, response.StatusCode)
		return
	}

	// Default fallback filename if we can't extract it.
	filename := "fallback_name.unknown"

	// Attempt to extract filename from Content-Disposition header.
	if cd := response.Header.Get("Content-Disposition"); cd != "" {
		re := regexp.MustCompile(`(?i)filename="?([^";]+)"?`)
		if match := re.FindStringSubmatch(cd); len(match) == 2 {
			filename = strings.TrimSpace(match[1])
		}
	}

	// Clean and sanitize the filename: remove or replace unsafe characters.
	filename = sanitizeFilename(filename)

	// Create the full output file path.
	outPath := path.Join(workingDirectory, filename)

	// Check if the file already exists.
	if fileExists(outPath) {
		log.Printf("Already exists: %s\n", outPath)
		return
	}

	// Create the output file.
	fileOutput, err := os.Create(outPath)
	if err != nil {
		log.Printf("Error creating file %s: %v\n", outPath, err)
		return
	}
	defer fileOutput.Close() // Ensure the file is closed later.

	// Copy the content of the HTTP response directly into the file.
	_, err = io.Copy(fileOutput, response.Body)
	if err != nil {
		log.Printf("Error writing file %s: %v\n", outPath, err)
		return
	}

	// Log successful download.
	log.Printf("Downloaded: %s\n", outPath)
}

// sanitizeFilename replaces invalid characters in filenames with underscores.
func sanitizeFilename(name string) string {
	// Allow only a-z, A-Z, 0-9, dot, dash and underscore.
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	return re.ReplaceAllString(name, "_")
}

// createDirectory attempts to create the directory at the given path with the provided permissions.
func createDirectory(path string, permission os.FileMode) {
	err := os.Mkdir(path, permission)
	if err != nil {
		log.Fatalf("Failed to create directory %s: %v\n", path, err)
	}
}

// directoryExists checks if a directory exists at the given path.
func directoryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
