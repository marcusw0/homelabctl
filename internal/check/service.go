package check

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type ServiceResults struct {
	HTTP HTTPResults
	DNS  DNSResults
	TCP  TCPResults
	TLS  TLSResults
}

type Service struct {
	FQDN    string
	IP      string
	Port    int
	Timeout time.Duration
}

type serviceChecks struct {
	HTTP func(context.Context) (HTTPResults, error)
	DNS  func(context.Context) (DNSResults, error)
	TCP  func(context.Context) (TCPResults, error)
	TLS  func(context.Context) (TLSResults, error)
}

func (c *Service) Check(ctx context.Context) (ServiceResults, error) {
	target := net.JoinHostPort(c.IP, strconv.Itoa(c.Port))

	httpCheck := HTTP{
		Timeout:        c.Timeout,
		ExpectedStatus: http.StatusOK,
		FollowRedirect: true,
	}
	httpTarget := "https://" + net.JoinHostPort(
		c.FQDN,
		strconv.Itoa(c.Port),
	)
	dns := DNS{Timeout: c.Timeout}
	tcp := TCP{Timeout: c.Timeout}
	tls := TLS{Port: c.Port, Timeout: c.Timeout}

	return runServiceChecks(ctx, serviceChecks{
		HTTP: func(ctx context.Context) (HTTPResults, error) {
			return httpCheck.Check(ctx, httpTarget)
		},
		DNS: func(ctx context.Context) (DNSResults, error) {
			return dns.Check(ctx, c.FQDN)
		},
		TCP: func(ctx context.Context) (TCPResults, error) {
			return tcp.Check(ctx, target)
		},
		TLS: func(ctx context.Context) (TLSResults, error) {
			return tls.Check(ctx, c.FQDN)
		},
	})

}

func runServiceChecks(
	ctx context.Context,
	checks serviceChecks,
) (ServiceResults, error) {
	var (
		results ServiceResults
		httpErr error
		dnsErr  error
		tcpErr  error
		tlsErr  error
		wg      sync.WaitGroup
	)

	wg.Add(4)

	go func() {
		defer wg.Done()

		results.HTTP, httpErr = checks.HTTP(ctx)
		if httpErr != nil {
			httpErr = fmt.Errorf("HTTP check: %w", httpErr)
		}
	}()

	go func() {
		defer wg.Done()

		results.DNS, dnsErr = checks.DNS(ctx)
		if dnsErr != nil {
			dnsErr = fmt.Errorf("DNS check: %w", dnsErr)
		}
	}()

	go func() {
		defer wg.Done()

		results.TCP, tcpErr = checks.TCP(ctx)
		if tcpErr != nil {
			tcpErr = fmt.Errorf("TCP check: %w", tcpErr)
		}
	}()

	go func() {
		defer wg.Done()

		results.TLS, tlsErr = checks.TLS(ctx)
		if tlsErr != nil {
			tlsErr = fmt.Errorf("TLS check: %w", tlsErr)
		}
	}()

	wg.Wait()

	return results, errors.Join(httpErr, dnsErr, tcpErr, tlsErr)
}

func (r ServiceResults) Healthy() bool {
	return r.HTTP.Healthy &&
		r.DNS.Healthy &&
		r.TCP.Healthy &&
		r.TLS.Healthy
}
