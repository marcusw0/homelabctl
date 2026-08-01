package check

import (
	"context"
	"fmt"
	"net"
	"time"
)

type TCPResults struct {
	Target    string
	Healthy   bool
	Latency   time.Duration
	CheckedAt time.Time
}

var ERROR_FAILED_DIAL = fmt.Errorf("Failed to dial:")

func CheckTCP(ctx context.Context, target string) (TCPResults, error) {
	var d net.Dialer
	ctx, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()

	start := time.Now()
	conn, err := d.DialContext(ctx, "tcp", target)
	if err != nil {
		results := TCPResults{
			Target:    target,
			Healthy:   false,
			Latency:   time.Since(start),
			CheckedAt: time.Now(),
		}

		return results, err
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
