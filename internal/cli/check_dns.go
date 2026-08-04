package cli

import (
	"context"
	"errors"
	"flag"
	"log"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
)

type DNSCheckCommand struct {
	Target  string
	Timeout time.Duration
}

func parseDNSCheck(args []string) (Command, error) {
	cmd := &DNSCheckCommand{}

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

func (c *DNSCheckCommand) Validate() error {
	if c.Target == "" {
		return errors.New("Need to specify destination to check")
	}

	if c.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	return nil
}

func (c *DNSCheckCommand) Run(ctx context.Context) error {
	service := check.DNS{
		Timeout: c.Timeout,
	}

	resp, err :=  service.Check(ctx, c.Target)
	if err != nil {
		log.Printf(
			"Host: %s\nResponse: %v\nLatency: %dms\nHealthy: %t\nChecked At: %v\n",
			resp.Target,
			resp.Response,
			resp.Latency.Milliseconds(),
			resp.Healthy,
			resp.CheckedAt.Local(),
		)
		return err
	}

	handleDNSResponse(resp)
	return nil
}

func handleDNSResponse(resp check.DNSResults) {
	log.Printf(
		"Host: %s\nResponse: %v\nLatency: %dms\nHealthy: %t\nChecked At: %v\n",
		resp.Target,
		resp.Response,
		resp.Latency.Milliseconds(),
		resp.Healthy,
		resp.CheckedAt.Local(),
	)
}
