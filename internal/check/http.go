package check

import (
	"context"
	"net/http"
	"time"
)

type HTTP struct {
	Timeout        time.Duration
	ExpectedStatus int
	FollowRedirect bool
}

type HTTPResults struct {
	Target     string
	Status     string
	StatusCode int
	Healthy    bool
	Latency    time.Duration
	CheckedAt  time.Time
	Body       string
}

func (h *HTTP) Check(
	ctx context.Context,
	target string,
) (HTTPResults, error) {

	ctx, cancel := context.WithTimeout(ctx, h.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", target, nil)
	if err != nil {
		result := HTTPResults{
			Target:    target,
			Healthy:   false,
			CheckedAt: time.Now(),
		}
		return result, err
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 {
				if !h.FollowRedirect {
					return http.ErrUseLastResponse
				}
			}
			return nil
		},
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		result := HTTPResults{
			Target:    target,
			Latency:   time.Since(start),
			Healthy:   false,
			CheckedAt: time.Now(),
		}
		return result, err
	}

	defer resp.Body.Close()

	result := HTTPResults{
		Target:     target,
		StatusCode: resp.StatusCode,
		Latency:    time.Since(start),
		Healthy:    resp.StatusCode == h.ExpectedStatus,
		CheckedAt:  time.Now(),
	}

	if resp.StatusCode >= 400 {
		result.Healthy = false
	}

	return result, nil
}
