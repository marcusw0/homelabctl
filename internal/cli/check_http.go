package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/marcusw0/homelabctl/internal/check"
)

func runHTTPCheck(ctx context.Context, request HTTPCheckRequest, streams IOStreams) error {
	timeout := defaultTimeout
	if request.Timeout != nil {
		timeout = *request.Timeout
	}
	if timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}
	if request.ExpectedStatus < 100 || request.ExpectedStatus > 599 {
		return errors.New("expected-status must be between 100 and 599")
	}
	target, err := check.NormalizeHTTPURL(request.Target)
	if err != nil {
		return err
	}

	checkHTTP := check.HTTP{
		Timeout:        timeout,
		ExpectedStatus: request.ExpectedStatus,
		FollowRedirect: request.FollowRedirect,
	}

	resp, respErr := checkHTTP.Check(ctx, target)
	writeErr := writeHTTPResponse(streams.Out, resp, request.Verbose)
	if writeErr != nil {
		return errors.Join(respErr, writeErr)
	}
	return respErr
}

func writeHTTPResponse(
	out io.Writer,
	resp check.HTTPResults,
	verbose bool,
) error {
	if !verbose {
		if resp.StatusCode == 0 {
			_, err := fmt.Fprintf(
				out,
				"Host: %s\nHealthy: %t\n",
				resp.Target,
				resp.Healthy,
			)
			return err
		}

		_, err := fmt.Fprintf(
			out,
			"Host: %s\nStatus: %d\nHealthy: %t\n",
			resp.Target,
			resp.StatusCode,
			resp.Healthy,
		)
		return err
	}

	if resp.StatusCode == 0 {
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
			"Status: %d\n"+
			"Latency: %s\n"+
			"Healthy: %t\n"+
			"Checked At: %s\n",
		resp.Target,
		resp.StatusCode,
		formatDuration(resp.Latency),
		resp.Healthy,
		formatTimestamp(resp.CheckedAt),
	)
	if err != nil {
		return err
	}
	return nil
}
