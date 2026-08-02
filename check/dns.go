package check

import (
	"context"
	"net"
	"time"
)

type DNSResults struct {
	Target    string
	Response  []string
	Healthy   bool
	Latency   time.Duration
	CheckedAt time.Time
}

func CheckDNS(ctx context.Context, target string) (DNSResults, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
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
