package logger

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"
)

// FileDestination writes logs to a file
type FileDestination struct {
	file *os.File
}

// NewFileDestination initializes a new file destination for logs
func NewFileDestination(filePath string) (*FileDestination, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("could not open log file: %w", err)
	}
	return &FileDestination{file: file}, nil
}

// Write writes log entry to a file
func (f *FileDestination) Write(p []byte) (n int, err error) {
	return f.file.Write(p)
}

// Close closes the file
func (f *FileDestination) Close() error {
	return f.file.Close()
}

// ConsoleDestination writes logs to the console
type ConsoleDestination struct{}

// Write writes log entry to console (stdout)
func (c *ConsoleDestination) Write(p []byte) (n int, err error) {
	return fmt.Fprintf(os.Stdout, "%s", p)
}

// Close closes the console destination (no-op for console)
func (c *ConsoleDestination) Close() error {
	return nil
}

// RemoteDestination simulates sending logs to a remote aggregator (e.g., Prometheus or Grafana)
type RemoteDestination struct {
	url string

	// client with Transport that reuses connections
	client *http.Client
}

func NewRemoteDestination(url string) *RemoteDestination {

	// Create a custom HTTP client with a Transport that reuses connections
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,               // Maximum idle connections across all hosts
			MaxIdleConnsPerHost: 5,                // Max idle connections to a single host
			IdleConnTimeout:     90 * time.Second, // Keep-alive timeout
		},
	}

	return &RemoteDestination{url: url, client: client}
}

// Write sends log entry to remote destination
func (r *RemoteDestination) Write(p []byte) (n int, err error) {
	req, err := http.NewRequest("POST", r.url, bytes.NewBuffer(p))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return len(p), nil
}

// Close closes the remote destination (no-op for this example)
func (r *RemoteDestination) Close() error {
	return nil
}
