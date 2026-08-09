package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
)

type DNSCheckCmd struct {
	Target  string
	Timeout time.Duration
}

func parseDNSCheck(args []string) (Command, error) {
	cmd := &DNSCheckCmd{}

	flags := flag.NewFlagSet("check dns", flag.ContinueOnError)

	flags.DurationVar(
		&cmd.Timeout,
		"timeout",
		5*time.Second,
		"timeout duration",
	)

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	if flags.NArg() == 1 {
		cmd.Target = flags.Arg(0)
	}

	if flags.NArg() > 1 {
		return nil, errors.New("check dns accepts only one argument")
	}

	return cmd, nil
}

func (c *DNSCheckCmd) Validate() error {
	if c.Target == "" {
		return errors.New("Need to specify destination to check")
	}

	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	return nil
}

func (c *DNSCheckCmd) Run(ctx context.Context, streams IOStreams) error {
	out := streams.Out
	service := check.DNS{
		Timeout: c.Timeout,
	}

	resp, respErr := service.Check(ctx, c.Target)
	if respErr != nil {
		_, writeErr := fmt.Fprintf(
			out,
			"Host: %s\nResponse: %v\nLatency: %dms\nHealthy: %t\nChecked At: %v\n",
			resp.Target,
			resp.Response,
			resp.Latency.Milliseconds(),
			resp.Healthy,
			resp.CheckedAt.Local(),
		)
		return errors.Join(respErr, writeErr)
	}

	if err := handleDNSResponse(out, resp); err != nil {
		return err
	}
	return nil
}

func handleDNSResponse(out io.Writer, resp check.DNSResults) error {
	if _, err := fmt.Fprintf(
		out,
		"Host: %s\nResponse: %v\nLatency: %dms\nHealthy: %t\nChecked At: %v\n",
		resp.Target,
		resp.Response,
		resp.Latency.Milliseconds(),
		resp.Healthy,
		resp.CheckedAt.Local(),
	); err != nil {
		return err
	}
	return nil
}
