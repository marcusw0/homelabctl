package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
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
	if checkErr == nil && !results.Healthy() {
		checkErr = fmt.Errorf("service %q is unhealthy", c.serviceName)
	}
	writeErr := writeService(
		streams.Out,
		results,
		c.serviceName,
		c.verbose,
	)

	return errors.Join(checkErr, ctx.Err(), writeErr)
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

	var output strings.Builder
	fmt.Fprintln(&output, serviceName)
	fmt.Fprintln(&output, "\x1b[31mHTTP RESULTS\x1b[0m")
	if resp.HTTP.Status != check.StatusSkipped {
		fmt.Fprintf(&output, "Status: %d\nLatency: %s\n", resp.HTTP.Result.StatusCode, formatDuration(resp.HTTP.Result.Latency))
	}
	fmt.Fprintf(&output, "Health: %s\n-----------\n", resp.HTTP.Status)
	fmt.Fprintln(&output, "\x1b[31mDNS RESULTS\x1b[0m")
	if resp.DNS.Status != check.StatusSkipped {
		fmt.Fprintf(&output, "Response: %v\nLatency: %s\n", resp.DNS.Result.Response, formatDuration(resp.DNS.Result.Latency))
	}
	fmt.Fprintf(&output, "Health: %s\n-----------\n", resp.DNS.Status)
	fmt.Fprintln(&output, "\x1b[31mTCP RESULTS\x1b[0m")
	if resp.TCP.Status != check.StatusSkipped {
		fmt.Fprintf(&output, "Latency: %s\nMessage: %q\n", formatDuration(resp.TCP.Result.Latency), resp.TCP.Result.Message)
	}
	fmt.Fprintf(&output, "Health: %s\n-----------\n", resp.TCP.Status)
	fmt.Fprintln(&output, "\x1b[31mTLS RESULTS\x1b[0m")
	if resp.TLS.Status != check.StatusSkipped {
		fmt.Fprintf(&output, "Subject: %s\nIssuer: %s\nName: %s\nNotAfter: %s\nExpires: %s\n",
			resp.TLS.Result.Subject, resp.TLS.Result.Issuer, resp.TLS.Result.Names,
			formatTimestamp(resp.TLS.Result.After), formatExpiry(resp.TLS.Result.Expires))
	}
	fmt.Fprintf(&output, "Health: %s\n-----------\n", resp.TLS.Status)
	_, err := io.WriteString(out, output.String())
	return err
}
