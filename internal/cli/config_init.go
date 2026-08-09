package cli

import (
	"context"
	"errors"

	"github.com/marcusw0/homelabctl/internal/config"
)

type ConfigInitCmd struct {
	ConfigPath string
}

func (c *ConfigInitCmd) Validate() error {
	if c.ConfigPath == "" {
		return errors.New("config path cannot be empty")
	}
	return nil
}

func (c *ConfigInitCmd) Run(ctx context.Context) error {
	if err := config.Initialize(c.ConfigPath); err != nil {
		return err
	}

	return nil
}
