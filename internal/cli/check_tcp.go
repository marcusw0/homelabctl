package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/marcusw0/homelabctl/internal/check"
)

func runTCPCheck(
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
	if err := check.ValidateAddrPort(request.Target); err != nil {
		return err
	}

	service := check.TCP{
		Timeout: timeout,
	}

	resp, respErr := service.Check(ctx, request.Target)
	writeErr := writeTCPResponse(streams.Out, resp, request.Verbose)
	if writeErr != nil {
		return errors.Join(respErr, writeErr)
	}
	return respErr
}

func writeTCPResponse(
	out io.Writer,
	resp check.TCPResults,
	verbose bool,
) error {
	if !verbose {
		_, err := fmt.Fprintf(
			out,
			"Host: %s\nHealthy: %t\nMessage: %q\n",
			resp.Target,
			resp.Healthy,
			resp.Message,
		)
		return err
	}

	_, err := fmt.Fprintf(
		out,
		"Host: %s\n"+
			"Latency: %s\n"+
			"Healthy: %t\n"+
			"Checked At: %s\n"+
			"Message: %q\n",
		resp.Target,
		formatDuration(resp.Latency),
		resp.Healthy,
		formatTimestamp(resp.CheckedAt),
		resp.Message,
	)
	if err != nil {
		return err
	}
	return nil
}
