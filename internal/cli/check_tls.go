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

type TLSCheckCmd struct {
	Target  string
	Port    int
	Timeout time.Duration
	Verbose bool
}

func parseTLSCheck(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Command, error) {
	cmd := &TLSCheckCmd{}

	flags := flag.NewFlagSet("check tls", flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	flags.IntVar(
		&cmd.Port,
		"port",
		443,
		"set TLS port",
	)

	flags.DurationVar(
		&cmd.Timeout,
		"timeout",
		3*time.Second,
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
		return nil, errors.New("check tls accepts only one argument")
	}

	return cmd, nil
}

func (c *TLSCheckCmd) Validate() error {
	if c.Target == "" {
		return errors.New("Expected destination")
	}

	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("port number must be any number from 1-65535")
	}

	return nil
}

func (c *TLSCheckCmd) Run(ctx context.Context, streams IOStreams) error {
	service := check.TLS{
		Port:    c.Port,
		Timeout: c.Timeout,
	}

	resp, respErr := service.Check(ctx, c.Target)
	writeErr := writeTLSResponse(streams.Out, resp, c.Verbose)
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
