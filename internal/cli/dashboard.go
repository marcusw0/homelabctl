package cli

import (
	"context"
	"io"

	"github.com/marcusw0/homelabctl/internal/check"
	"github.com/marcusw0/homelabctl/internal/config"
	"github.com/marcusw0/homelabctl/internal/tui/dashboard"
)

type DashboardCmd struct {
	cfgPath    string
	serviceCfg map[string]config.Service
	checks     map[string]check.Service
	names      []string
}

func parseDashboard(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Command, error) {
	cmd := DashboardCmd{
		cfgPath:    opts.ConfigPath,
		serviceCfg: make(map[string]config.Service),
		checks:     make(map[string]check.Service),
	}
	return cmd, nil
}

func (d DashboardCmd) Validate() error {
	path := d.cfgPath
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	for name, service := range cfg.Services {
		if service.Enabled {
			d.serviceCfg[name] = service
		}
	}
	return nil
}

func (d DashboardCmd) Run(ctx context.Context, streams IOStreams) error {
	for name, service := range d.serviceCfg {
		d.names = append(d.names, name)
		d.checks[name] = serviceFromConfig(service)
	}
	if err := dashboard.Run(ctx, d.names, d.checks); err != nil {
		return err
	}
	return nil
}
