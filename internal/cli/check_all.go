package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/marcusw0/homelabctl/internal/check"
	"github.com/marcusw0/homelabctl/internal/config"
	"github.com/marcusw0/homelabctl/internal/runner"
)

type AllCmd struct {
	ConfigPath string
	Servers    map[string]config.Server
}

type serviceResult struct {
	name   string
	result check.ServiceResults
	err    error
}

func (c *AllCmd) Validate() error {
	cfg, err := config.Load(c.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	servers := make(map[string]config.Server)
	for name, server := range cfg.Servers {
		if server.Enabled {
			servers[name] = server
		}
	}
	if len(servers) == 0 {
		return errors.New("no enabled services configured")
	}

	c.Servers = servers

	return nil
}

func (c *AllCmd) Run(ctx context.Context, streams IOStreams) error {
	const maxConcurrent = 5

	jobs := make([]runner.Job, 0, len(c.Servers))

	for name, service := range c.Servers {
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

	for result := range serviceRunner.Run(ctx, jobs) {
		results[result.Name] = result
	}

	return writeAllResults(streams.ErrOut, streams.Out, results)
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
