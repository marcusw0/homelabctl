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

type TCPCheckCmd struct {
	Target  string
	Port    int
	Timeout time.Duration
	Verbose bool
}

func parseTCPCheck(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Command, error) {
	cmd := &TCPCheckCmd{}

	flags := flag.NewFlagSet("check tcp", flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	flags.IntVar(
		&cmd.Port,
		"port",
		443,
		"port number",
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
		return nil, errors.New("check tcp accepts only one argument")
	}

	return cmd, nil
}

func (c *TCPCheckCmd) Validate() error {
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

func (c *TCPCheckCmd) Run(ctx context.Context, streams IOStreams) error {
	service := check.TCP{
		Port:    c.Port,
		Timeout: c.Timeout,
	}

	resp, respErr := service.Check(ctx, c.Target)
	writeErr := writeTCPResponse(streams.Out, resp, c.Verbose)
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
