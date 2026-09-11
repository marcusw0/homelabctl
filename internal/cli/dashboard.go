package cli

import (
	"context"
	"io"

	"github.com/marcusw0/homelabctl/internal/check"
	"github.com/marcusw0/homelabctl/internal/config"
	"github.com/marcusw0/homelabctl/internal/tui/dashboard"
)

type DashboardCmd struct {
	cfgPath string
	checks map[string][]check.Kind
	names []string
}

func parseDashboardCmd(
	errOut io.Writer,
	args []string,
	opts GlobalOption,
) (Command, error) {
	cmd := DashboardCmd{
		cfgPath: opts.ConfigPath,
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
			d.names = append(d.names, name)
			d.checks[name] = service.Checks
		}
	}
	return nil
}

func (d DashboardCmd) Run(ctx context.Context, streams IOStreams) error {
	if err := dashboard.Run(ctx, d.names, d.checks); err != nil {
		return err
	}
	return nil
}
