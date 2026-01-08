// Package internal provides the core functionality for the file downloader,
// including download management, progress tracking, and utility functions.
package internal

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
)

// validateURL checks if the given URL is in a proper format using url.Parse.
// Only http and https schemes are allowed.
func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme '%s': only http and https are supported", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("invalid URL: missing host")
	}

	return nil
}

// ensureDirectory checks if the directory exists and creates it if it doesn't.
// Returns an error if the path exists but is not a directory, or if creation fails.
func ensureDirectory(dir string) error {
	if dir == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	info, err := os.Stat(dir)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path exists but is not a directory: %s", dir)
		}
		return nil
	}

	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		return nil
	}

	return err
}

// extractFileName extracts the filename from a URL by taking the last path segment
// and removing any query parameters.
func extractFileName(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Get the path and extract the last segment
	path := parsedURL.Path
	segments := strings.Split(path, "/")
	filename := segments[len(segments)-1]

	if filename == "" {
		return "download", nil
	}

	return filename, nil
}

// safeClose safely closes an io.Closer and logs any error that occurs.
func safeClose(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("error closing resource: %v", err)
	}
}
