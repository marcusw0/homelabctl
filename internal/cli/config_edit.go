package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/marcusw0/homelabctl/internal/config"
	"github.com/marcusw0/homelabctl/internal/tui/serviceform"
)

type ConfigEditCmd struct {
	ConfigPath  string
	ServiceName string
	Config config.Service
}

func (c *ConfigEditCmd) Validate() error {
	if c.ConfigPath == "" {
		return errors.New("config path cannot be empty")
	}
	if c.ServiceName == "" {
		return errors.New("specify a name for the new service")
	}
	cfg, err := config.Load(c.ConfigPath)
	if err != nil {
		return err
	}

	existing, ok := cfg.Services[c.ServiceName]
	if !ok {
		return fmt.Errorf("service %q is not in config", c.ServiceName)
	}
	c.Config = existing

	return nil
}

func (c *ConfigEditCmd) Run(ctx context.Context, streams IOStreams) error {
	newService, err := serviceform.Run(
		ctx,
		streams.In,
		streams.ErrOut,
		c.ServiceName,
		&c.Config,
	)
	if err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	newServiceV2 := fillDefaults(newService)

	if err := config.UpdateService(
		c.ConfigPath,
		c.ServiceName,
		newServiceV2,
	); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		streams.Out,
		"changes to %q have been saved",
		c.ServiceName,
	); err != nil {
		return err
	}
	return nil
}

