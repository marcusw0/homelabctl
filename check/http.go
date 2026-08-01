package check

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPResult contains the useful parts of an HTTP health check.
type HTTPResult struct {
	StatusCode int
	Latency    time.Duration
	Healthy    bool
	Body       string
}

// CheckHTTP requests target and reports whether it returned a 2xx status.
func CheckHTTP(ctx context.Context, target string) (HTTPResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := http.NewRequestWithContext(ctx, "GET", target, nil)
	if err != nil {
		return HTTPResult{}, err
	}

	start := time.Now()
	resp, err := http.DefaultClient.Do(client)
	if err != nil {
		return HTTPResult{}, fmt.Errorf("check %q: %w", target, err)
	}

	defer resp.Body.Close()

	result := HTTPResult{
		StatusCode: resp.StatusCode,
		Latency:    time.Since(start),
		Healthy:    resp.StatusCode >= 200 && resp.StatusCode < 300,
	}

	if !result.Healthy {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return HTTPResult{}, fmt.Errorf("read response from %q: %w", target, err)
		}
		result.Body = string(body)
	}

	return result, nil
}
