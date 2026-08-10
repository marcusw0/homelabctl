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

func parseTLSCheck(errOut io.Writer, args []string, opts GlobalOption) (Command, error) {
	cmd := &TLSCheckCmd{}

	flags := flag.NewFlagSet("check tls", flag.ContinueOnError)
	flags.SetOutput(errOut)

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

	flags.BoolVar(
		&cmd.Verbose,
		"verbose",
		false,
		"verbose output",
	)

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

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
		Verbose: c.Verbose,
	}

	resp, respErr := service.Check(ctx, c.Target)
	if respErr != nil {
		_, writeErr := fmt.Fprintf(
			streams.Out,
			"Host: %s\nHealthy: %t\nChecked At: %v\n",
			resp.Target,
			resp.Healthy,
			resp.CheckedAt.Local(),
		)
		return errors.Join(respErr, writeErr)
	}

	if err := writeTLSResponse(streams.Out, resp); err != nil {
		return err
	}
	return nil
}

func writeTLSResponse(out io.Writer, resp check.TLSResults) error {
	_, err := fmt.Fprintf(
		out,
		"Host: %s\nSubject: %s\nIssuer: %s\nName: %s\nNotAfter: %s\nExpires: %v\nLatency: %dms\nHealthy: %t\nChecked At: %v\n",
		resp.Target,
		resp.Subject,
		resp.Issuer,
		resp.Names,
		resp.After,
		formatExpiry(resp.Expires),
		resp.Latency.Milliseconds(),
		resp.Healthy,
		resp.CheckedAt.Local(),
	)
	if err != nil {
		return err
	}
	return nil
}

func formatExpiry(d time.Duration) string {
	if d < 0 {
		return "EXPIRED"
	}

	days := int(d / (24 * time.Hour))
	hours := int((d % (24 * time.Hour)) / time.Hour)

	return fmt.Sprintf("%dd %dh", days, hours)
}
