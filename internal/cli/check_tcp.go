package cli

import (
	"context"
	"errors"
	"flag"
	"log"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
)

type TCPCheckCommand struct {
	Target  string
	Port    int
	Timeout time.Duration
}

func parseTCPCheck(args []string) (Command, error) {
	cmd := &TCPCheckCommand{}

	flags := flag.NewFlagSet("check tcp", flag.ContinueOnError)

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

	if flags.NArg() == 1 {
		cmd.Target = flags.Arg(0)
	}

	if flags.NArg() > 1 {
		return nil, errors.New("check tcp accepts only one argument")
	}

	return cmd, nil
}

func (c *TCPCheckCommand) Validate() error {
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

func (c *TCPCheckCommand) Run(ctx context.Context) error {
	service := check.TCP{
		Port:    c.Port,
		Timeout: c.Timeout,
	}

	resp, err := service.Check(ctx, c.Target)
	if err != nil {
		log.Printf(
			"Host: %s\nLatency: %dms\nHealthy: %t\nChecked At: %v\nMessage: %q\n",
			resp.Target,
			resp.Latency.Milliseconds(),
			resp.Healthy,
			resp.CheckedAt.Local(),
			resp.Message,
		)
		return err
	}

	handleTCPResponse(resp)
	return nil
}

func handleTCPResponse(resp check.TCPResults) {
	log.Printf(
		"Host: %s\nLatency: %dms\nHealthy: %t\nChecked At: %v\nMessage: %q\n",
		resp.Target,
		resp.Latency.Milliseconds(),
		resp.Healthy,
		resp.CheckedAt.Local(),
		resp.Message,
	)
}
