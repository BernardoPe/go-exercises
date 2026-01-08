package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid http URL",
			url:     "http://example.com/file.txt",
			wantErr: false,
		},
		{
			name:    "valid https URL",
			url:     "https://example.com/file.txt",
			wantErr: false,
		},
		{
			name:    "valid URL with path",
			url:     "https://example.com/path/to/file.txt",
			wantErr: false,
		},
		{
			name:    "valid URL with query params",
			url:     "https://example.com/file.txt?param=value",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid protocol",
			url:     "ftp://example.com/file.txt",
			wantErr: true,
		},
		{
			name:    "no protocol",
			url:     "example.com/file.txt",
			wantErr: true,
		},
		{
			name:    "invalid format",
			url:     "not a url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractFileName(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "simple filename",
			url:      "https://example.com/file.txt",
			expected: "file.txt",
		},
		{
			name:     "filename with path",
			url:      "https://example.com/path/to/file.txt",
			expected: "file.txt",
		},
		{
			name:     "filename with query params",
			url:      "https://example.com/file.txt?param=value",
			expected: "file.txt",
		},
		{
			name:     "filename with multiple query params",
			url:      "https://example.com/file.txt?param1=value1&param2=value2",
			expected: "file.txt",
		},
		{
			name:     "filename without extension",
			url:      "https://example.com/download",
			expected: "download",
		},
		{
			name:     "numeric filename",
			url:      "https://httpbin.org/bytes/10000000",
			expected: "10000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractFileName(tt.url)
			if err != nil {
				t.Errorf("extractFileName() error = %v", err)
				return
			}
			if got != tt.expected {
				t.Errorf("extractFileName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEnsureDirectory(t *testing.T) {
	tests := []struct {
		name      string
		dir       string
		wantErr   bool
		cleanupFn func()
	}{
		{
			name:    "empty directory path",
			dir:     "",
			wantErr: true,
		},
		{
			name:    "create new directory",
			dir:     filepath.Join(t.TempDir(), "test-dir"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cleanupFn != nil {
				defer tt.cleanupFn()
			}

			err := ensureDirectory(tt.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureDirectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.dir != "" {
				info, err := os.Stat(tt.dir)
				if err != nil {
					t.Errorf("directory was not created: %v", err)
					return
				}
				if !info.IsDir() {
					t.Error("created path is not a directory")
				}
			}
		})
	}
}

func TestEnsureDirectoryExisting(t *testing.T) {
	tempDir := t.TempDir()
	existingDir := filepath.Join(tempDir, "existing")
	err := os.Mkdir(existingDir, os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	err = ensureDirectory(existingDir)
	if err != nil {
		t.Errorf("ensureDirectory() on existing directory failed: %v", err)
	}
}
