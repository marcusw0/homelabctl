package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/marcusw0/homelabctl/internal/check"
	"github.com/marcusw0/homelabctl/internal/config"
)

func runServiceCheck(
	ctx context.Context,
	request CheckRequest,
	path string,
	streams IOStreams,
) error {
	if request.Timeout != nil && *request.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	cfg, err := config.Load(path)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	serviceCfg, exists := cfg.Services[request.Target]
	if !exists {
		return fmt.Errorf(
			"service %q not found in %s",
			request.Target,
			path,
		)
	}
	if !serviceCfg.Enabled {
		return fmt.Errorf("service %q is disabled", request.Target)
	}
	if request.Timeout != nil {
		serviceCfg.Timeout = *request.Timeout
	}

	service := serviceFromConfig(serviceCfg)

	results, checkErr := service.Check(ctx)
	writeErr := writeService(
		streams.Out,
		results,
		request.Target,
		request.Verbose,
	)

	return errors.Join(checkErr, ctx.Err(), writeErr)
}

func writeService(
	out io.Writer,
	resp check.ServiceResults,
	serviceName string,
	verbose bool,
) error {
	if !verbose {
		_, err := fmt.Fprintf(
			out,
			"HTTP status: %s\n"+
				"DNS status: %s\n"+
				"TCP status: %s\n"+
				"TLS status: %s\n",
			resp.HTTP.Status,
			resp.DNS.Status,
			resp.TCP.Status,
			resp.TLS.Status,
		)
		return err
	}

	var output strings.Builder
	fmt.Fprintln(&output, serviceName)
	fmt.Fprintln(&output, "\x1b[31mHTTP RESULTS\x1b[0m")
	if resp.HTTP.Status != check.StatusSkipped {
		fmt.Fprintf(&output, "Status: %d\nLatency: %s\n", resp.HTTP.Result.StatusCode, formatDuration(resp.HTTP.Result.Latency))
	}
	fmt.Fprintf(&output, "Health: %s\n-----------\n", resp.HTTP.Status)
	fmt.Fprintln(&output, "\x1b[31mDNS RESULTS\x1b[0m")
	if resp.DNS.Status != check.StatusSkipped {
		fmt.Fprintf(&output, "Response: %v\nLatency: %s\n", resp.DNS.Result.Response, formatDuration(resp.DNS.Result.Latency))
	}
	fmt.Fprintf(&output, "Health: %s\n-----------\n", resp.DNS.Status)
	fmt.Fprintln(&output, "\x1b[31mTCP RESULTS\x1b[0m")
	if resp.TCP.Status != check.StatusSkipped {
		fmt.Fprintf(&output, "Latency: %s\nMessage: %q\n", formatDuration(resp.TCP.Result.Latency), resp.TCP.Result.Message)
	}
	fmt.Fprintf(&output, "Health: %s\n-----------\n", resp.TCP.Status)
	fmt.Fprintln(&output, "\x1b[31mTLS RESULTS\x1b[0m")
	if resp.TLS.Status != check.StatusSkipped {
		fmt.Fprintf(&output, "Subject: %s\nIssuer: %s\nName: %s\nNotAfter: %s\nExpires: %s\n",
			resp.TLS.Result.Subject, resp.TLS.Result.Issuer, resp.TLS.Result.Names,
			formatTimestamp(resp.TLS.Result.After), formatExpiry(resp.TLS.Result.Expires))
	}
	fmt.Fprintf(&output, "Health: %s\n-----------\n", resp.TLS.Status)
	_, err := io.WriteString(out, output.String())
	return err
}

func serviceFromConfig(service config.Service) check.Service {
	return check.Service{
		FQDN:            service.FQDN,
		TCPHost:         service.EffectiveTCPHost(),
		Port:            service.Port,
		Timeout:         service.EffectiveTimeout(),
		TLSWarnBefore:   service.EffectiveTLSWarnBefore(),
		Checks:          service.EffectiveChecks(),
		HTTPURL:         service.EffectiveHTTPURL(),
		ExpectedStatus:  service.EffectiveStatusCode(),
		FollowRedirects: service.EffectiveFollowRedirects(),
	}
}
