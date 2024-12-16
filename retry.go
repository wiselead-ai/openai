package openaicli

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"time"
)

const (
	maxRetries     = 5
	baseRetryDelay = 2 * time.Second
	maxRetryDelay  = 30 * time.Second
)

func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := c.httpClient.Do(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		delay := time.Duration(float64(baseRetryDelay) * math.Pow(2, float64(attempt)))
		if delay > maxRetryDelay {
			delay = maxRetryDelay
		}

		select {
		case <-req.Context().Done():
			return nil, fmt.Errorf("request cancelled or timed out: %w", req.Context().Err())
		case <-time.After(delay):
			continue
		}
	}
	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func isTemporaryError(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}
	return true
}
