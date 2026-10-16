package cli

import (
	"context"
	"slices"

	"github.com/marcusw0/homelabctl/internal/check"
	"github.com/marcusw0/homelabctl/internal/config"
	"github.com/marcusw0/homelabctl/internal/tui/dashboard"
)

func runDashboard(ctx context.Context, path string) error {
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	names := make([]string, 0, len(cfg.Services))
	checks := make(map[string]check.Service)

	for name, service := range cfg.Services {
		if service.Enabled {
			names = append(names, name)
			checks[name] = serviceFromConfig(service)
		}
	}
	slices.Sort(names)
	if err := dashboard.Run(ctx, names, checks); err != nil {
		return err
	}
	return nil
}
