package cli

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
	"github.com/marcusw0/homelabctl/internal/config"
	"github.com/marcusw0/homelabctl/internal/runner"
)

type AllCmd struct {
	ConfigPath string
	Servers    map[string]ServiceCheckCmd
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
	servers := make(map[string]ServiceCheckCmd)
	for k, v := range cfg.Servers {
		if v.Enabled == true {
			servers[k] = ServiceCheckCmd{
				fqdn:    v.FQDN,
				ip:      v.IP,
				port:    v.Port,
				timeout: 5 * time.Second,
			}
		}
		continue
	}
	if len(servers) == 0 {
		return fmt.Errorf("No services configured.")
	}

	c.Servers = servers

	return nil

}

func (c *AllCmd) Run(ctx context.Context, streams IOStreams) error {
	const maxConcurrent = 5

	jobs := make([]runner.Job, 0, len(c.Servers))

	for name, server := range c.Servers {
		service := check.Service{
			FQDN:    server.fqdn,
			IP:      server.ip,
			Port:    server.port,
			Timeout: server.timeout,
		}

		jobs = append(jobs, runner.Job{
			Name:    name,
			Checker: &service,
		})
	}

	serviceRunner := runner.Runner{
		MaxConcurrent: 5,
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
			"%s\t%t\t%t\t%t\t%t\n",
			k,
			v.Checks.HTTP.Healthy,
			v.Checks.DNS.Healthy,
			v.Checks.TCP.Healthy,
			v.Checks.TLS.Healthy,
		); err != nil {
			return err
		}
	}
	return writer.Flush()
}
