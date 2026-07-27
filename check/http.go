package check

import (
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
func CheckHTTP(target string) (HTTPResult, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	start := time.Now()
	resp, err := client.Get(target)
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
