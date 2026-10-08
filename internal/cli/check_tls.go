package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/marcusw0/homelabctl/internal/check"
)

func runTLSCheck(
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

	target, port, err := parseTLSTarget(request.Target)
	if err != nil {
		return err
	}

	service := check.TLS{
		Port: port,
		Timeout: timeout,
	}

	resp, respErr := service.Check(ctx, target)
	writeErr := writeTLSResponse(streams.Out, resp, request.Verbose)
	if writeErr != nil {
		return errors.Join(respErr, writeErr)
	}
	return respErr
}

func writeTLSResponse(
	out io.Writer,
	resp check.TLSResults,
	verbose bool,
) error {
	if !verbose {
		_, err := fmt.Fprintf(
			out,
			"Host: %s\nHealthy: %t\n",
			resp.Target,
			resp.Healthy,
		)
		return err
	}

	if resp.After.IsZero() {
		_, err := fmt.Fprintf(
			out,
			"Host: %s\nHealthy: %t\nChecked At: %s\n",
			resp.Target,
			resp.Healthy,
			formatTimestamp(resp.CheckedAt),
		)
		return err
	}

	_, err := fmt.Fprintf(
		out,
		"Host: %s\n"+
			"Subject: %s\n"+
			"Issuer: %s\n"+
			"Name: %s\n"+
			"NotAfter: %s\n"+
			"Expires: %v\n"+
			"Latency: %s\n"+
			"Healthy: %t\n"+
			"Checked At: %s\n",
		resp.Target,
		resp.Subject,
		resp.Issuer,
		resp.Names,
		formatTimestamp(resp.After),
		formatExpiry(resp.Expires),
		formatDuration(resp.Latency),
		resp.Healthy,
		formatTimestamp(resp.CheckedAt),
	)
	if err != nil {
		return err
	}
	return nil
}
