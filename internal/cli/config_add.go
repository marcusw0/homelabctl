package cli

import (
	"context"
	"errors"

	"github.com/marcusw0/homelabctl/internal/config"
)

type ConfigAddCmd struct {
	ConfigPath string
	ServerName string
}

func (c *ConfigAddCmd) Validate() error {
	if c.ConfigPath == "" {
		return errors.New("config path cannot be empty")
	}
	if c.ServerName == "" {
		return errors.New("specify a name for the new server")
	}
	return nil
}

func (c *ConfigAddCmd) Run(ctx context.Context) error {
	if err := config.ConfigAdd(c.ServerName, c.ConfigPath); err != nil {
		return err
	}

	return nil
}
