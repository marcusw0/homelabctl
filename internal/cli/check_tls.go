package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
)

type TLSCheckCommand struct {
	Target  string
	Port    int
	Timeout time.Duration
	Verbose bool
}

func parseTLSCheck(args []string) (Command, error) {
	cmd := &TLSCheckCommand{}

	flags := flag.NewFlagSet("check tls", flag.ContinueOnError)

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

func (c *TLSCheckCommand) Validate() error {
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

func (c *TLSCheckCommand) Run(ctx context.Context) error {
	service := check.TLS{
		Port:    c.Port,
		Timeout: c.Timeout,
		Verbose: c.Verbose,
	}

	resp, err :=  service.Check(ctx, c.Target)
	if err != nil {
		log.Printf(
			"Host: %s\nHealthy: %t\nChecked At: %v\n",
			resp.Target,
			resp.Healthy,
			resp.CheckedAt.Local(),
		)

		return err
	}

	handleTLSResponse(resp)
	return nil
}

func handleTLSResponse(resp check.TLSResults) {
	log.Printf(
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
}

func formatExpiry(d time.Duration) string {
	if d < 0 {
		return "EXPIRED"
	}

	days := int(d / (24 * time.Hour))
	hours := int((d % (24 * time.Hour)) / time.Hour)

	return fmt.Sprintf("%dd %dh", days, hours)
}

