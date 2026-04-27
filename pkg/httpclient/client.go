package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client with retry logic
type Client struct {
	httpClient  *http.Client
	maxRetries  int
	backoffBase time.Duration
}

// NewClient creates a new HTTP client
func NewClient(timeout time.Duration, maxRetries int) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries:  maxRetries,
		backoffBase: 1 * time.Second,
	}
}

// GetWithRetry performs an HTTP GET request with retry logic
func (c *Client) GetWithRetry(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := c.backoffBase * time.Duration(1<<uint(attempt-1))
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		// Set headers
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = err
			continue
		}

		// Retry on 5xx errors
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			continue
		}

		// Don't retry on 4xx errors
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("client error: %d", resp.StatusCode)
		}

		return body, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
