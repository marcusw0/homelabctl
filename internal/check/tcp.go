package check

import (
	"context"
	"net"
	"strconv"
	"time"
)

type TCP struct {
	Port    int
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

	port := strconv.Itoa(c.Port)
	host := net.JoinHostPort(target, port)

	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", host)
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
