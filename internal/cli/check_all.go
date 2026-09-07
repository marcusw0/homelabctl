package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/marcusw0/homelabctl/internal/config"
	"github.com/marcusw0/homelabctl/internal/runner"
)

type AllCmd struct {
	ConfigPath string
	Services   map[string]config.Service
}

func (c *AllCmd) Validate() error {
	cfg, err := config.Load(c.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	services := make(map[string]config.Service)
	for name, service := range cfg.Services {
		if service.Enabled {
			services[name] = service
		}
	}
	if len(services) == 0 {
		return errors.New("no enabled services configured")
	}

	c.Services = services

	return nil
}

func (c *AllCmd) Run(ctx context.Context, streams IOStreams) error {
	const maxConcurrent = 5

	jobs := make([]runner.Job, 0, len(c.Services))

	for name, service := range c.Services {
		s := serviceFromConfig(service)

		jobs = append(jobs, runner.Job{
			Name:    name,
			Checker: &s,
		})
	}

	serviceRunner := runner.Runner{
		MaxConcurrent: maxConcurrent,
	}

	results := make(map[string]runner.Result)
	var errs []error

	for result := range serviceRunner.Run(ctx, jobs) {
		results[result.Name] = result
		if result.Err != nil {
			errs = append(errs, fmt.Errorf("service %q: %w", result.Name, result.Err))
		} else if !result.Checks.Healthy() {
			errs = append(errs, fmt.Errorf("service %q is unhealthy", result.Name))
		}
	}

	errs = append(errs, ctx.Err(), writeAllResults(streams.ErrOut, streams.Out, results))
	return errors.Join(errs...)
}

func writeAllResults(errOut io.Writer, out io.Writer, results map[string]runner.Result) error {
	writer := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(
		writer,
		"NAME\tHTTP\tDNS\tTCP\tTLS",
	); err != nil {
		return err
	}

	for k, v := range results {
		if v.Err != nil {
			fmt.Fprintf(errOut, "[ERROR]%s returned: %v\n", k, v.Err)
		}
		if _, err := fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%s\t%s\n",
			k,
			v.Checks.HTTP.Status,
			v.Checks.DNS.Status,
			v.Checks.TCP.Status,
			v.Checks.TLS.Status,
		); err != nil {
			return err
		}
	}
	return writer.Flush()
}
