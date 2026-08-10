package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
	"github.com/marcusw0/homelabctl/internal/config"
)

type CheckServiceCmd struct {
	timeout time.Duration
	fqdn    string
	ip      string
	port    int
}

func parseCheckService(errOut io.Writer, args []string, opts GlobalOption) (Command, error) {
	cmd := &CheckServiceCmd{}

	flags := flag.NewFlagSet("check service", flag.ContinueOnError)
	flags.SetOutput(errOut)

	flags.DurationVar(
		&cmd.timeout,
		"timeout",
		5*time.Second,
		"Request timeout",
	)

	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	if flags.NArg() != 1 {
		return nil, errors.New("check service accepts exactly one server name")
	}

	serverName := flags.Arg(0)

	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("invalid config %w", err)
	}

	server, exists := cfg.Servers[serverName]
	if !exists {
		return nil, fmt.Errorf(
			"server %q not found in %s",
			serverName,
			opts.ConfigPath,
		)
	}
	if !server.Enabled {
		return nil, fmt.Errorf("server %q is disabled", serverName)
	}

	checkService := CheckServiceCmd{
		timeout: cmd.timeout,
		fqdn:    server.FQDN,
		ip:      server.IP,
		port:    server.Port,
	}
	return &checkService, nil
}

func (c *CheckServiceCmd) Validate() error {
	if c.timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	if c.fqdn == "" {
		return errors.New("fqdn cannot be blank. Check config")
	}

	if c.ip == "" {
		return errors.New("ip cannot be blank. Check config")
	}

	if c.port <= 0 || c.port > 65535 {
		return errors.New("port number must be any number from 1-65535")
	}
	return nil
}

func (c *CheckServiceCmd) Run(ctx context.Context, streams IOStreams) error {
	service := check.CheckService{
		FQDN:    c.fqdn,
		IP:      c.ip,
		Port:    c.port,
		Timeout: c.timeout,
	}
	results, checkErr := service.Check(ctx)
	writeErr := writeService(streams.Out, results)
	if writeErr != nil {
		return errors.Join(checkErr, writeErr)
	}
	return checkErr
}

func writeService(out io.Writer, results check.CheckServiceResults) error {
	_, err := fmt.Fprintf(
		out,
		"HTTP healthy: %t\nDNS healthy: %t\nTCP healthy: %t\nTLS healthy: %t\n",
		results.HTTP.Healthy,
		results.DNS.Healthy,
		results.TCP.Healthy,
		results.TLS.Healthy,
	)
	if err != nil {
		return err
	}
	return nil
}
