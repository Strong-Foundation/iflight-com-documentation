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
	"sync"
	"time"
)

func main() {
	// Define the working directory where files will be saved.
	workingDirectory := "assets/"

	// Check if the working directory exists; if not, create it with permission 0700.
	if !directoryExists(workingDirectory) {
		createDirectory(workingDirectory, 0700)
	}

	// Loop from 0 up to the maximum download ID (loopStopCounter).
	// You can increase this if needed.
	loopStopCounter := 100000

	// Create a waitgroup.
	var downloadWaitGroup sync.WaitGroup

	// Iterate through the download IDs.
	for index := 0; index <= loopStopCounter; index++ {
		// Sleep for 1 second before each request.
		time.Sleep(1 * time.Second)
		// Add a 1 to the counter.
		downloadWaitGroup.Add(1)
		// Construct the download URL dynamically.
		url := fmt.Sprintf("https://shop.iflight.com/index.php?route=product/product/download&download_id=%d", index)
		// Download the file for this URL.
		go downloadFiles(url, workingDirectory, &downloadWaitGroup)
	}
	downloadWaitGroup.Wait()
}

// downloadFiles sends an HTTP GET request to the given URL, saves the response body to disk,
// and skips files that already exist or have an invalid response.
// It uses a 60-second timeout and a WaitGroup to manage concurrency.
func downloadFiles(givenURL string, workingDirectory string, waitGroup *sync.WaitGroup) {
	// Notify the WaitGroup when the function exits.
	defer waitGroup.Done()

	// Create an HTTP client with a timeout.
	client := http.Client{
		Timeout: 30 * time.Minute,
	}

	// Send a GET request to the given URL using the client with timeout.
	response, err := client.Get(givenURL)
	if err != nil {
		// Log and return if there is an error making the request.
		log.Printf("Error making request to %s: %v\n", givenURL, err)
		return
	}
	// Ensure the response body is closed when the function returns.
	defer response.Body.Close()
	// Skip this file if the server does not return HTTP 200 OK or the content is empty.
	if response.StatusCode != http.StatusOK || response.ContentLength == 0 {
		log.Printf("Skipping: %s - StatusCode: %d\n", givenURL, response.StatusCode)
		return
	}
	// Set a default filename in case we can't determine one from the headers.
	filename := "fallback_name.unknown"
	// Try to extract the filename from the Content-Disposition header if available.
	if cd := response.Header.Get("Content-Disposition"); cd != "" {
		// Regular expression to find the filename in the Content-Disposition header.
		re := regexp.MustCompile(`(?i)filename="?([^";]+)"?`)
		if match := re.FindStringSubmatch(cd); len(match) == 2 {
			// Trim spaces and use the extracted filename.
			filename = strings.TrimSpace(match[1])
		}
	}
	// Sanitize the filename by replacing unsafe characters with underscores, etc.
	// Allow only a-z, A-Z, 0-9, dot, dash and underscore.
	regexStringChanger := regexp.MustCompile(`[^a-zA-Z0-9.]`)
	// Change the file name using regex.
	filename = regexStringChanger.ReplaceAllString(filename, "_")
	// Collapse multiple underscores
	filename = regexp.MustCompile(`_+`).ReplaceAllString(filename, "_")
	// Build the full path where the file will be saved.
	outPath := path.Join(workingDirectory, filename)
	// Lower the string and the file name.
	outPath = strings.ToLower(outPath)
	// Check if the file already exists to avoid re-downloading it.
	if fileExists(outPath) {
		log.Printf("Already exists: %s\n", outPath)
		return
	}
	// Attempt to create the file for writing.
	fileOutput, err := os.Create(outPath)
	if err != nil {
		// Log an error if the file cannot be created.
		log.Printf("Error creating file %s: %v\n", outPath, err)
		return
	}
	// Ensure the file is closed when the function exits.
	defer fileOutput.Close()
	// Stream the response body directly into the output file.
	_, err = io.Copy(fileOutput, response.Body)
	if err != nil {
		// Log any error that occurs while writing to the file.
		log.Printf("Error writing file %s: %v\n", outPath, err)
		return
	}
	// Log a message to indicate successful download.
	log.Printf("Downloaded: %s\n", outPath)
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
