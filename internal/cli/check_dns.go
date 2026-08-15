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
	Verbose bool
}

func parseDNSCheck(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Command, error) {

	cmd := &DNSCheckCmd{}

	flags := flag.NewFlagSet("check dns", flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	flags.DurationVar(
		&cmd.Timeout,
		"timeout",
		5*time.Second,
		"timeout duration",
	)

	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	cmd.Verbose = opts.Verbose

	if flags.NArg() == 1 {
		cmd.Target = flags.Arg(0)
	}

	if flags.NArg() > 1 {
		return nil, errors.New("check dns accepts only one argument")
	}

	return cmd, nil
}

func (c *DNSCheckCmd) Validate() error {
	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	if err := check.ValidateHostname(c.Target); err != nil {
		return err
	}

	return nil
}

func (c *DNSCheckCmd) Run(ctx context.Context, streams IOStreams) error {
	out := streams.Out
	service := check.DNS{
		Timeout: c.Timeout,
	}

	resp, respErr := service.Check(ctx, c.Target)
	writeErr := writeDNSResponse(out, resp, c.Verbose)
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
