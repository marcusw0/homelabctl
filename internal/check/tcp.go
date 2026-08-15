package check

import (
	"context"
	"net"
	"time"
)

type TCP struct {
	Timeout time.Duration
}

type TCPResults struct {
	Target    string
	Healthy   bool
	Latency   time.Duration
	CheckedAt time.Time
	Message   string
}

func (c *TCP) Check(ctx context.Context, target string) (TCPResults, error) {
	var d net.Dialer
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", target)
	if err != nil {
		results := TCPResults{
			Target:    target,
			Healthy:   false,
			Latency:   time.Since(start),
			CheckedAt: time.Now(),
			Message:   err.Error(),
		}
		// treat context cancellation as error
		if ctx.Err() != nil {
			return results, ctx.Err()
		}
		// treat unreachable as unhealthy, not as an error
		return results, nil
	}
	defer conn.Close()

	results := TCPResults{
		Target:    target,
		Healthy:   true,
		Latency:   time.Since(start),
		CheckedAt: time.Now(),
	}

	return results, nil
}
