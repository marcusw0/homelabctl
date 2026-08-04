package check

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTP struct {
	Timeout        time.Duration
	ExpectedStatus int
	FollowRedirect bool
}

// HTTPResult contains the useful parts of an HTTP health check.
type HTTPResult struct {
	Target     string
	Status     string
	StatusCode int
	Healthy    bool
	Latency    time.Duration
	CheckedAt  time.Time
	Body       string
}

func (h *HTTP) Check(ctx context.Context, target string) (HTTPResult, error) {
	ctx, cancel := context.WithTimeout(ctx, h.Timeout)
	defer cancel()
	client, err := http.NewRequestWithContext(ctx, "GET", target, nil)
	if err != nil {
		result := HTTPResult{
			StatusCode: client.Response.StatusCode,
			Healthy:    false,
			CheckedAt:  time.Now(),
		}
		return result, err
	}

	start := time.Now()
	resp, err := http.DefaultClient.Do(client)
	if err != nil {
		result := HTTPResult{
			Latency:   time.Since(start),
			Healthy:   false,
			CheckedAt: time.Now(),
		}
		return result, err
	}

	defer resp.Body.Close()

	result := HTTPResult{
		StatusCode: resp.StatusCode,
		Latency:    time.Since(start),
		Healthy:    resp.StatusCode >= 200 && resp.StatusCode < 300,
		CheckedAt:  time.Now(),
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
