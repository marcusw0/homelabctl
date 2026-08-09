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

type CheckServiceResults struct {
	HTTP HTTPResult
	DNS  DNSResults
	TCP  TCPResults
	TLS  TLSResults
}

type CheckService struct {
	FQDN string
	IP   string
	Port int
}

func (c *CheckService) Check(
	ctx context.Context,
) (CheckServiceResults, error) {

	var errs []error

	httpCheck := HTTP{
		Timeout:        5 * time.Second,
		ExpectedStatus: http.StatusOK,
		FollowRedirect: true,
	}
	httpTarget := "https://" + net.JoinHostPort(
		c.FQDN,
		strconv.Itoa(c.Port),
	)
	httpResp, err := httpCheck.Check(ctx, httpTarget)
	if err != nil {
		errs = append(errs, fmt.Errorf("HTTP check: %w\n", err))
	}

	dns := DNS{
		Timeout: 3 * time.Second,
	}
	dnsResp, err := dns.Check(ctx, c.FQDN)
	if err != nil {
		errs = append(errs, fmt.Errorf("DNS check: %w\n", err))
	}

	tcp := TCP{
		Port:    c.Port,
		Timeout: 3 * time.Second,
	}
	tcpResp, err := tcp.Check(ctx, c.IP)
	if err != nil {
		errs = append(errs, fmt.Errorf("TCP check: %w\n", err))
	}

	tls := TLS{
		Port:    c.Port,
		Timeout: 5 * time.Second,
	}
	tlsResp, err := tls.Check(ctx, c.FQDN)
	if err != nil {
		errs = append(errs, fmt.Errorf("TLS check: %w\n", err))
	}

	combined := CheckServiceResults{
		HTTP: httpResp,
		DNS:  dnsResp,
		TCP:  tcpResp,
		TLS:  tlsResp,
	}

	if len(errs) == 0 && !combined.Healthy() {
		errs = append(errs, errors.New("one or more checks are unhealthy"))
	}

	return combined, errors.Join(errs...)
}

func (r CheckServiceResults) Healthy() bool {
	return r.HTTP.Healthy &&
		r.DNS.Healthy &&
		r.TCP.Healthy &&
		r.TLS.Healthy
}
