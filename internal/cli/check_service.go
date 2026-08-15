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

type ServiceCheckCmd struct {
	server  string
	timeout time.Duration
	fqdn    string
	ip      string
	port    int
	verbose bool
}

func parseServiceCheck(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Command, error) {
	cmd := &ServiceCheckCmd{}

	flags := flag.NewFlagSet("check service", flag.ContinueOnError)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

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

	checkService := ServiceCheckCmd{
		server:  serverName,
		timeout: cmd.timeout,
		fqdn:    server.FQDN,
		ip:      server.IP,
		port:    server.Port,
		verbose: opts.Verbose,
	}
	return &checkService, nil
}

func (c *ServiceCheckCmd) Validate() error {
	if c.timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}

	if err := check.ValidateHostname(c.fqdn); err != nil {
		return err
	}

	if err := check.ValidateIP(c.ip); err != nil {
		return err
	}

	if err := check.ValidatePort(c.port); err != nil {
		return err
	}

	return nil
}

func (c *ServiceCheckCmd) Run(ctx context.Context, streams IOStreams) error {
	service := check.Service{
		FQDN:    c.fqdn,
		IP:      c.ip,
		Port:    c.port,
		Timeout: c.timeout,
	}

	results, checkErr := service.Check(ctx)
	writeErr := writeService(streams.Out, results, c.server, c.verbose)
	if writeErr != nil {
		return errors.Join(checkErr, writeErr)
	}
	return checkErr
}

func writeService(
	out io.Writer,
	results check.ServiceResults,
	server string,
	verbose bool,
) error {
	if !verbose {
		_, err := fmt.Fprintf(
			out,
			"HTTP healthy: %t\n"+
				"DNS healthy: %t\n"+
				"TCP healthy: %t\n"+
				"TLS healthy: %t\n",
			results.HTTP.Healthy,
			results.DNS.Healthy,
			results.TCP.Healthy,
			results.TLS.Healthy,
		)
		return err
	}

	_, err := fmt.Fprintf(
		out,
		"%s\n"+
			"\x1b[31mHTTP RESULTS\x1b[0m\n"+
			"Status: %d\n"+
			"Latency: %s\n"+
			"Healthy: %t\n"+
			"-----------\n"+
			"\x1b[31mDNS RESULTS\x1b[0m\n"+
			"Response: %v\n"+
			"Latency: %s\n"+
			"Healthy: %t\n"+
			"-----------\n"+
			"\x1b[31mTCP RESULTS\x1b[0m\n"+
			"Latency: %s\n"+
			"Healthy: %t\n"+
			"Message: %q\n"+
			"-----------\n"+
			"\x1b[31mTLS RESULTS\x1b[0m\n"+
			"Subject: %s\n"+
			"Issuer: %s\n"+
			"Name: %s\n"+
			"NotAfter: %s\n"+
			"Expires: %v\n"+
			"Healthy: %t\n"+
			"-----------\n",
		server,
		results.HTTP.StatusCode,
		formatDuration(results.HTTP.Latency),
		results.HTTP.Healthy,
		results.DNS.Response,
		formatDuration(results.DNS.Latency),
		results.DNS.Healthy,
		formatDuration(results.TCP.Latency),
		results.TCP.Healthy,
		results.TCP.Message,
		results.TLS.Subject,
		results.TLS.Issuer,
		results.TLS.Names[:],
		formatTimestamp(results.TLS.After),
		formatExpiry(results.TLS.Expires),
		results.TLS.Healthy,
	)
	if err != nil {
		return err
	}

	return nil
}
