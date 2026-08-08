package cli

import (
	"context"
	"errors"

	"github.com/marcusw0/homelabctl/internal/config"
)

func (c *ConfigInitCommand) Validate() error {
	if c.ConfigPath == "" {
		return errors.New("config path cannot be empty")
	}
	return nil
}

func (c *ConfigInitCommand) Run(ctx context.Context) error {
	if err := config.Initialize(c.ConfigPath); err != nil {
		return err
	}

	return nil
}
