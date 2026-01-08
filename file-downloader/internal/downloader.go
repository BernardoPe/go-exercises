package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const (
	// DefaultBufferSize is the buffer size used for reading from network and writing to disk.
	DefaultBufferSize = 32 * 1024
)

// FileDownload represents a single file download with its progress channel.
type FileDownload struct {
	URL         string
	FilePath    string
	LoadedBytes chan int64
	totalBytes  int64
	mu          sync.RWMutex
	err         error
}

// Err returns the error that occurred during download, if any.
func (f *FileDownload) Err() error {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.err
}

// setErr sets the error in a thread-safe manner.
func (f *FileDownload) setErr(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.err = err
}

// TotalBytes returns the total bytes for the download.
func (f *FileDownload) TotalBytes() int64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.totalBytes
}

// setTotalBytes sets the total bytes for the download in a thread-safe manner.
func (f *FileDownload) setTotalBytes(total int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.totalBytes = total
}

// Prepare validates the URL and prepares the file path, but does NOT make any HTTP requests.
// This allows for setup of UI to consume the LoadedBytes channel before data starts flowing.
// The actual HTTP request is deferred until Start() is called.
func (f *FileDownload) Prepare(url string, directory string) error {
	f.URL = url

	err := validateURL(url)
	if err != nil {
		return err
	}

	fileName, err := extractFileName(url)
	if err != nil {
		return err
	}

	err = ensureDirectory(directory)
	if err != nil {
		return err
	}

	f.FilePath = filepath.Join(directory, fileName)
	f.LoadedBytes = make(chan int64)

	return nil
}

// Start makes the HTTP request, begins reading from the response, writing to file,
// and sending progress updates. Accepts a context for cancellation support.
func (f *FileDownload) Start(ctx context.Context, wg *sync.WaitGroup) error {
	if f.URL == "" {
		return errors.New("download not prepared: URL is empty")
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(f.LoadedBytes)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.URL, nil)
		if err != nil {
			f.setErr(fmt.Errorf("failed to create request: %w", err))
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			f.setErr(fmt.Errorf("failed to fetch URL: %w", err))
			return
		}
		defer safeClose(resp.Body)

		if resp.StatusCode != http.StatusOK {
			f.setErr(fmt.Errorf("bad status: %s", resp.Status))
			return
		}

		f.setTotalBytes(resp.ContentLength)

		file, err := os.Create(f.FilePath)
		if err != nil {
			f.setErr(fmt.Errorf("failed to create file: %w", err))
			return
		}
		defer safeClose(file)

		buffer := make([]byte, DefaultBufferSize)
		for {
			select {
			case <-ctx.Done():
				f.setErr(ctx.Err())
				return
			default:
			}

			n, err := resp.Body.Read(buffer)
			if n > 0 {
				_, writeErr := file.Write(buffer[:n])
				if writeErr != nil {
					f.setErr(fmt.Errorf("failed to write to file: %w", writeErr))
					return
				}
				f.LoadedBytes <- int64(n)
			}
			if err != nil {
				if err != io.EOF {
					f.setErr(fmt.Errorf("failed to read response: %w", err))
				}
				return
			}
		}
	}()

	return nil
}

// PrepareDownloads prepares all downloads without starting them.
// It validates URLs and sets up file paths but does not make HTTP requests.
func PrepareDownloads(urls []string, directory string) ([]*FileDownload, error) {
	err := ensureDirectory(directory)
	if err != nil {
		return nil, err
	}

	downloads := make([]*FileDownload, 0, len(urls))
	for _, url := range urls {
		d := &FileDownload{}
		err := d.Prepare(url, directory)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare download for %s: %w", url, err)
		}
		downloads = append(downloads, d)
	}

	return downloads, nil
}

// StartAll starts all downloads with the provided context and WaitGroup.
// It returns immediately after starting all goroutines; use WaitGroup to wait for completion.
func StartAll(ctx context.Context, downloads []*FileDownload, wg *sync.WaitGroup) error {
	for _, d := range downloads {
		err := d.Start(ctx, wg)
		if err != nil {
			return fmt.Errorf("failed to start download for %s: %w", d.URL, err)
		}
	}
	return nil
}
