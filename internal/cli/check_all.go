package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
)

type CheckAllResults struct {
	HTTP check.HTTPResult
	DNS  check.DNSResults
	TCP  check.TCPResults
	TLS  check.TLSResults
}

func (r CheckAllResults) Healthy() bool {
	return r.HTTP.Healthy &&
		r.DNS.Healthy &&
		r.TCP.Healthy &&
		r.TLS.Healthy
}

type CheckAllCmd struct {
	fqdn string
	ip   string
	port int
}

func (c *CheckAllCmd) Validate() error {
	if c.fqdn == "" {
		return errors.New("fqdn cannot be blank. Check config")
	}

	if c.ip == "" {
		return errors.New("ip cannot be blank. Check config")
	}

	if c.port <= 0 || c.port > 65535 {
		return errors.New("port number must be any number from 1-65535")
	}
	return nil
}

func (c *CheckAllCmd) Run(ctx context.Context) error {
	results, err := c.check(ctx)
	allResults(results)
	return err
}

func (c *CheckAllCmd) check(
	ctx context.Context,
) (CheckAllResults, error) {

	var errs []error

	httpCheck := check.HTTP{
		Timeout:        5 * time.Second,
		ExpectedStatus: http.StatusOK,
		FollowRedirect: true,
	}
	httpTarget := "https://" + net.JoinHostPort(
		c.fqdn,
		strconv.Itoa(c.port),
	)
	httpResp, err := httpCheck.Check(ctx, httpTarget)
	if err != nil {
		errs = append(errs, fmt.Errorf("HTTP check: %w", err))
	}

	dns := check.DNS{
		Timeout: 3 * time.Second,
	}
	dnsResp, err := dns.Check(ctx, c.fqdn)
	if err != nil {
		errs = append(errs, fmt.Errorf("DNS check: %w", err))
	}

	tcp := check.TCP{
		Port:    c.port,
		Timeout: 3 * time.Second,
	}
	tcpResp, err := tcp.Check(ctx, c.ip)
	if err != nil {
		errs = append(errs, fmt.Errorf("TCP check: %w", err))
	}

	tls := check.TLS{
		Port:    c.port,
		Timeout: 5 * time.Second,
	}
	tlsResp, err := tls.Check(ctx, c.fqdn)
	if err != nil {
		errs = append(errs, fmt.Errorf("TLS check: %w", err))
	}

	combined := CheckAllResults{
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

func allResults(results CheckAllResults) {
	fmt.Printf("HTTP healthy: %t\n", results.HTTP.Healthy)
	fmt.Printf("DNS healthy:  %t\n", results.DNS.Healthy)
	fmt.Printf("TCP healthy:  %t\n", results.TCP.Healthy)
	fmt.Printf("TLS healthy:  %t\n", results.TLS.Healthy)
}
