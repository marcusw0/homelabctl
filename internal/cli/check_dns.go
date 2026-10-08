package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/marcusw0/homelabctl/internal/check"
)

func runDNSCheck(
	ctx context.Context,
	request CheckRequest,
	streams IOStreams,
) error {
	timeout := defaultTimeout
	if request.Timeout != nil {
		timeout = *request.Timeout
	}
	if timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}
	if err := check.ValidateHostname(request.Target); err != nil {
		return err
	}

	checkDNS := check.DNS{
		Timeout: timeout,
	}

	resp, respErr := checkDNS.Check(ctx, request.Target)
	writeErr := writeDNSResponse(streams.Out, resp, request.Verbose)
	if writeErr != nil {
		return errors.Join(respErr, writeErr)
	}
	return respErr
}

func writeDNSResponse(
	out io.Writer,
	resp check.DNSResults,
	verbose bool,
) error {
	if !verbose {
		_, err := fmt.Fprintf(
			out,
			"Host: %s\nResponse: %v\nHealthy: %t\n",
			resp.Target,
			resp.Response,
			resp.Healthy,
		)
		return err
	}

	if _, err := fmt.Fprintf(
		out,
		"Host: %s\n"+
			"Response: %v\n"+
			"Latency: %s\n"+
			"Healthy: %t\n"+
			"Checked At: %s\n",
		resp.Target,
		resp.Response,
		formatDuration(resp.Latency),
		resp.Healthy,
		formatTimestamp(resp.CheckedAt),
	); err != nil {
		return err
	}
	return nil
}
