package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

const httpTimeout = 1 * time.Second

type Observer interface {
	WriteEvent(*Event) error
}

// FileObserver appends each event as a JSON line to the file.
type FileObserver struct {
	mu   sync.Mutex
	path string
}

func NewFileObserver(path string) *FileObserver {
	return &FileObserver{
		path: path,
	}
}

func (f *FileObserver) WriteEvent(event *Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	if _, err = file.Write(data); err != nil {
		file.Close()
		return err
	}

	return file.Close()
}

// HTTPObserver sends each event with a POST request to the URL.
type HTTPObserver struct {
	url    string
	client *http.Client
}

func NewHTTPObserver(url string) *HTTPObserver {
	return &HTTPObserver{
		url:    url,
		client: &http.Client{Timeout: httpTimeout},
	}
}

func (h *HTTPObserver) WriteEvent(event *Event) error {
	encodedEvent, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, h.url, bytes.NewReader(encodedEvent))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body) // Let the connection be reused.

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("audit server %s responded with status %d", h.url, resp.StatusCode)
	}

	return nil
}
