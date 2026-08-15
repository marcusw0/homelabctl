package cli

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
	"github.com/marcusw0/homelabctl/internal/config"
)

type AllCmd struct {
	ConfigPath string
	Servers    map[string]ServiceCheckCmd
}

type AllResults struct {
	httpHealth bool
	tlsHealth  bool
	tcpHealth  bool
	dnsHealth  bool
	err        error
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

	results := make(map[string]AllResults)

	limit := make(chan struct{}, maxConcurrent)
	workCh := make(chan serviceResult, len(c.Servers))

	for name, server := range c.Servers {
		limit <- struct{}{}

		go func(name string, server ServiceCheckCmd) {
			defer func() {
				<-limit
			}()

			service := check.Service{
				FQDN:    server.fqdn,
				IP:      server.ip,
				Port:    server.port,
				Timeout: server.timeout,
			}
			result, err := service.Check(ctx)

			workCh <- serviceResult{
				name:   name,
				result: result,
				err:    err,
			}
		}(name, server)
	}

	for range len(c.Servers) {
		msg := <-workCh
		results[msg.name] = AllResults{
			httpHealth: msg.result.HTTP.Healthy,
			tlsHealth:  msg.result.TLS.Healthy,
			tcpHealth:  msg.result.TCP.Healthy,
			dnsHealth:  msg.result.DNS.Healthy,
			err:        msg.err,
		}
	}
	err := writeAllResults(streams.ErrOut, streams.Out, results)
	if err != nil {
		return err
	}
	return nil
}

func writeAllResults(errOut io.Writer, out io.Writer, results map[string]AllResults) error {
	writer := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(
		writer,
		"NAME\tHTTP\tDNS\tTCP\tTLS",
	); err != nil {
		return err
	}

	for k, v := range results {
		if v.err != nil {
			fmt.Fprintf(errOut, "[ERROR]%s returned: %v\n", k, v.err)
		}
		if _, err := fmt.Fprintf(
			writer,
			"%s\t%t\t%t\t%t\t%t\n",
			k,
			v.httpHealth,
			v.dnsHealth,
			v.tcpHealth,
			v.tlsHealth,
		); err != nil {
			return err
		}
	}
	return writer.Flush()
}
