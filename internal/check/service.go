package check

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
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

	var errs []error

	httpCheck := HTTP{
		Timeout:        c.Timeout,
		ExpectedStatus: http.StatusOK,
		FollowRedirect: true,
	}
	httpTarget := "https://" + net.JoinHostPort(
		c.FQDN,
		strconv.Itoa(c.Port),
	)
	httpResp, err := httpCheck.Check(ctx, httpTarget)
	if err != nil {
		errs = append(errs, fmt.Errorf("HTTP check: %w", err))
	}

	dns := DNS{
		Timeout: c.Timeout,
	}
	dnsResp, err := dns.Check(ctx, c.FQDN)
	if err != nil {
		errs = append(errs, fmt.Errorf("DNS check: %w", err))
	}

	tcp := TCP{
		Port:    c.Port,
		Timeout: c.Timeout,
	}
	tcpResp, err := tcp.Check(ctx, c.IP)
	if err != nil {
		errs = append(errs, fmt.Errorf("TCP check: %w", err))
	}

	tls := TLS{
		Port:    c.Port,
		Timeout: c.Timeout,
	}
	tlsResp, err := tls.Check(ctx, c.FQDN)
	if err != nil {
		errs = append(errs, fmt.Errorf("TLS check: %w", err))
	}

	combined := ServiceResults{
		HTTP: httpResp,
		DNS:  dnsResp,
		TCP:  tcpResp,
		TLS:  tlsResp,
	}

	return combined, errors.Join(errs...)
}

func (r ServiceResults) Healthy() bool {
	return r.HTTP.Healthy &&
		r.DNS.Healthy &&
		r.TCP.Healthy &&
		r.TLS.Healthy
}
