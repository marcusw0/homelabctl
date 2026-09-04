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
	configPath      string
	serviceName     string
	serviceCfg      config.Service
	timeoutOverride *time.Duration
	verbose         bool
}

func parseServiceCheck(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Command, error) {
	cmd := &ServiceCheckCmd{}

	flags := flag.NewFlagSet(
		"check service",
		flag.ContinueOnError,
	)
	flags.SetOutput(errOut)
	addGlobalFlags(flags, &opts)

	flags.Func(
		"timeout",
		"override configured timeout",
		func(value string) error {
			timeout, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid timeout: %w", err)
			}

			cmd.timeoutOverride = &timeout
			return nil
		},
	)

	if err := flags.Parse(args); err != nil {
		return nil, err
	}
	if flags.NArg() != 1 {
		return nil, errors.New(
			"check service accepts exactly one service name",
		)
	}

	cmd.configPath = opts.ConfigPath
	cmd.serviceName = flags.Arg(0)
	cmd.verbose = opts.Verbose

	return cmd, nil
}

func (c *ServiceCheckCmd) Validate() error {
	if c.timeoutOverride != nil && *c.timeoutOverride <= 0 {
		return errors.New("timeout must be greater than zero")
	}

	cfg, err := config.Load(c.configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	service, exists := cfg.Services[c.serviceName]
	if !exists {
		return fmt.Errorf(
			"service %q not found in %s",
			c.serviceName,
			c.configPath,
		)
	}
	if !service.Enabled {
		return fmt.Errorf("service %q is disabled", c.serviceName)
	}

	c.serviceCfg = service
	return nil
}

func serviceFromConfig(service config.Service) check.Service {
	return check.Service{
		FQDN:            service.FQDN,
		TCPHost:         service.EffectiveTCPHost(),
		Port:            service.Port,
		Timeout:         service.EffectiveTimeout(),
		TLSWarnBefore:   service.EffectiveTLSWarnBefore(),
		Checks:          service.EffectiveChecks(),
		HTTPURL:         service.EffectiveHTTPURL(),
		ExpectedStatus:  service.EffectiveStatusCode(),
		FollowRedirects: service.EffectiveFollowRedirects(),
	}
}

func (c *ServiceCheckCmd) Run(ctx context.Context, streams IOStreams) error {
	service := serviceFromConfig(c.serviceCfg)

	if c.timeoutOverride != nil {
		service.Timeout = *c.timeoutOverride
	}

	results, checkErr := service.Check(ctx)
	writeErr := writeService(
		streams.Out,
		results,
		c.serviceName,
		c.verbose,
	)

	return errors.Join(checkErr, writeErr)
}

func writeService(
	out io.Writer,
	resp check.ServiceResults,
	serviceName string,
	verbose bool,
) error {
	if !verbose {
		_, err := fmt.Fprintf(
			out,
			"HTTP status: %s\n"+
				"DNS status: %s\n"+
				"TCP status: %s\n"+
				"TLS status: %s\n",
			resp.HTTP.Status,
			resp.DNS.Status,
			resp.TCP.Status,
			resp.TLS.Status,
		)
		return err
	}

	_, err := fmt.Fprintf(
		out,
		"%s\n"+
			"\x1b[31mHTTP RESULTS\x1b[0m\n"+
			"Status: %d\n"+
			"Latency: %s\n"+
			"Health: %s\n"+
			"-----------\n"+
			"\x1b[31mDNS RESULTS\x1b[0m\n"+
			"Response: %v\n"+
			"Latency: %s\n"+
			"Health: %s\n"+
			"-----------\n"+
			"\x1b[31mTCP RESULTS\x1b[0m\n"+
			"Latency: %s\n"+
			"Health: %s\n"+
			"Message: %q\n"+
			"-----------\n"+
			"\x1b[31mTLS RESULTS\x1b[0m\n"+
			"Subject: %s\n"+
			"Issuer: %s\n"+
			"Name: %s\n"+
			"NotAfter: %s\n"+
			"Expires: %v\n"+
			"Health: %s\n"+
			"-----------\n",
		serviceName,
		resp.HTTP.Result.StatusCode,
		formatDuration(resp.HTTP.Result.Latency),
		resp.HTTP.Status,
		resp.DNS.Result.Response,
		formatDuration(resp.DNS.Result.Latency),
		resp.DNS.Status,
		formatDuration(resp.TCP.Result.Latency),
		resp.TCP.Status,
		resp.TCP.Result.Message,
		resp.TLS.Result.Subject,
		resp.TLS.Result.Issuer,
		resp.TLS.Result.Names[:],
		formatTimestamp(resp.TLS.Result.After),
		formatExpiry(resp.TLS.Result.Expires),
		resp.TLS.Status,
	)

	return err
}
