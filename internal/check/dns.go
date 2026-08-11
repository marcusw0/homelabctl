package check

import (
	"context"
	"net"
	"time"
)

type DNS struct {
	Timeout time.Duration
}

type DNSResults struct {
	Target    string
	Response  []string
	Healthy   bool
	Latency   time.Duration
	CheckedAt time.Time
}

func (d *DNS) Check(ctx context.Context, target string) (DNSResults, error) {

	ctx, cancel := context.WithTimeout(ctx, d.Timeout)
	defer cancel()

	start := time.Now()
	dnsTest, err := net.DefaultResolver.LookupHost(ctx, target)
	if err != nil {
		result := DNSResults{
			Target:    target,
			Response:  dnsTest,
			Healthy:   false,
			Latency:   time.Since(start),
			CheckedAt: time.Now(),
		}

		return result, err
	}

	result := DNSResults{
		Target:    target,
		Response:  dnsTest,
		Healthy:   true,
		Latency:   time.Since(start),
		CheckedAt: time.Now(),
	}

	return result, nil
}
