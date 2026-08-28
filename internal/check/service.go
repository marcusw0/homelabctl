package check

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type ServiceResults struct {
	HTTP Outcome[HTTPResults]
	DNS  Outcome[DNSResults]
	TCP  Outcome[TCPResults]
	TLS  Outcome[TLSResults]
}

type Service struct {
	FQDN            string
	TCPHost         string
	Port            int
	Checks          []Kind
	HTTPURL         string
	ExpectedStatus  int
	FollowRedirects bool
	Timeout         time.Duration
	TLSWarnBefore   time.Duration
}

type serviceChecks struct {
	HTTP func(context.Context) (HTTPResults, error)
	DNS  func(context.Context) (DNSResults, error)
	TCP  func(context.Context) (TCPResults, error)
	TLS  func(context.Context) (TLSResults, error)
}

func (c *Service) Check(ctx context.Context) (ServiceResults, error) {
	var srvChecks serviceChecks
	http := HTTP{
		Timeout:        c.Timeout,
		ExpectedStatus: c.ExpectedStatus,
		FollowRedirect: c.FollowRedirects,
	}
	dns := DNS{Timeout: c.Timeout}
	tcp := TCP{Timeout: c.Timeout}
	tls := TLS{
		Port:       c.Port,
		Timeout:    c.Timeout,
		WarnBefore: c.TLSWarnBefore,
	}

	for _, kind := range c.Checks {
		switch kind {
		case KindHTTP:
			srvChecks.HTTP = func(ctx context.Context) (HTTPResults, error) {
				return http.Check(ctx, c.HTTPURL)
			}
		case KindDNS:
			srvChecks.DNS = func(ctx context.Context) (DNSResults, error) {
				return dns.Check(ctx, c.FQDN)
			}
		case KindTCP:
			srvChecks.TCP = func(ctx context.Context) (TCPResults, error) {
				return tcp.Check(ctx, c.TCPHost)
			}
		case KindTLS:
			srvChecks.TLS = func(ctx context.Context) (TLSResults, error) {
				return tls.Check(ctx, c.FQDN)
			}
		}
	}

	return runServiceChecks(ctx, srvChecks)
}

func runServiceChecks(
	ctx context.Context,
	checks serviceChecks,
) (ServiceResults, error) {
	results := ServiceResults{
		HTTP: Outcome[HTTPResults]{Status: StatusSkipped},
		DNS:  Outcome[DNSResults]{Status: StatusSkipped},
		TCP:  Outcome[TCPResults]{Status: StatusSkipped},
		TLS:  Outcome[TLSResults]{Status: StatusSkipped},
	}

	var wg sync.WaitGroup

	if checks.HTTP != nil {
		wg.Go(func() {
			result, err := checks.HTTP(ctx)
			if err != nil {
				err = fmt.Errorf("HTTP check: %w", err)
			}
			results.HTTP = Outcome[HTTPResults]{
				Result: result,
				Err:    err,
				Status: statusFor(result.Healthy, err),
			}
		})
	}

	if checks.DNS != nil {
		wg.Go(func() {
			result, err := checks.DNS(ctx)
			if err != nil {
				err = fmt.Errorf("DNS check: %w", err)
			}
			results.DNS = Outcome[DNSResults]{
				Result: result,
				Err:    err,
				Status: statusFor(result.Healthy, err),
			}
		})
	}

	if checks.TCP != nil {
		wg.Go(func() {
			result, err := checks.TCP(ctx)
			if err != nil {
				err = fmt.Errorf("TCP check: %w", err)
			}
			results.TCP = Outcome[TCPResults]{
				Result: result,
				Err:    err,
				Status: statusFor(result.Healthy, err),
			}
		})
	}

	if checks.TLS != nil {
		wg.Go(func() {
			result, err := checks.TLS(ctx)
			if err != nil {
				err = fmt.Errorf("TLS check: %w", err)
			}
			status := statusFor(result.Healthy, err)

			if status == StatusHealthy && result.ExpiresSoon {
				status = StatusWarning
			}
			results.TLS = Outcome[TLSResults]{
				Result: result,
				Err:    err,
				Status: status,
			}
		})
	}
	wg.Wait()

	return results, errors.Join(
		results.HTTP.Err,
		results.DNS.Err,
		results.TCP.Err,
		results.TLS.Err,
	)
}

func (r ServiceResults) Healthy() bool {
	return outcomeHealthy(r.HTTP.Status) &&
		outcomeHealthy(r.DNS.Status) &&
		outcomeHealthy(r.TCP.Status) &&
		outcomeHealthy(r.TLS.Status)
}

func outcomeHealthy(status Status) bool {
	return status == StatusHealthy ||
		status == StatusWarning ||
		status == StatusSkipped
}

func statusFor(healthy bool, err error) Status {
	if err != nil {
		return StatusError
	}

	if healthy {
		return StatusHealthy
	}

	return StatusUnhealthy
}
