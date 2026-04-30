package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Payload represents the JSON body sent to the webhook endpoint.
type Payload struct {
	Process   string    `json:"process"`
	Event     string    `json:"event"`
	PID       int       `json:"pid,omitempty"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Sender dispatches alert payloads to a webhook URL.
type Sender struct {
	URL    string
	Client *http.Client
}

// NewSender creates a Sender with a sensible default HTTP client timeout.
func NewSender(url string) *Sender {
	return &Sender{
		URL: url,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send marshals p and POSTs it to the configured webhook URL.
// It returns an error if the request fails or the server responds with a
// non-2xx status code. The response body is included in the error message
// when a non-2xx status is received, to aid debugging.
func (s *Sender) Send(p Payload) error {
	if p.Timestamp.IsZero() {
		p.Timestamp = time.Now().UTC()
	}

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("alert: marshal payload: %w", err)
	}

	resp, err := s.Client.Post(s.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("alert: post webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		if len(respBody) > 0 {
			return fmt.Errorf("alert: webhook returned status %d: %s", resp.StatusCode, respBody)
		}
		return fmt.Errorf("alert: webhook returned status %d", resp.StatusCode)
	}
	return nil
}
