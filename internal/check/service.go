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

func (c *Service) Check(ctx context.Context) (ServiceResults, error) {
	var (
		results ServiceResults
		httpErr error
		dnsErr  error
		tcpErr  error
		tlsErr  error
		wg      sync.WaitGroup
	)

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
	dns := DNS{
		Timeout: c.Timeout,
	}
	tcp := TCP{
		Timeout: c.Timeout,
	}
	tls := TLS{
		Port:    c.Port,
		Timeout: c.Timeout,
	}

	wg.Add(4)
	go func() {
		defer wg.Done()
		results.HTTP, httpErr = httpCheck.Check(ctx, httpTarget)
		if httpErr != nil {
			httpErr = fmt.Errorf("HTTP check: %w", httpErr)
		}
	}()
	go func() {
		defer wg.Done()
		results.DNS, dnsErr = dns.Check(ctx, c.FQDN)
		if dnsErr != nil {
			dnsErr = fmt.Errorf("DNS check: %w", dnsErr)
		}
	}()
	go func() {
		defer wg.Done()
		results.TCP, tcpErr = tcp.Check(ctx, target)
		if tcpErr != nil {
			tcpErr = fmt.Errorf("TCP check: %w", tcpErr)
		}
	}()
	go func() {
		defer wg.Done()
		results.TLS, tlsErr = tls.Check(ctx, c.FQDN)
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
