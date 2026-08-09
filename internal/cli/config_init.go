package cli

import (
	"context"
	"errors"
	"fmt"

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

func (c *ConfigInitCmd) Run(ctx context.Context, streams IOStreams) error {
	if err := config.Initialize(c.ConfigPath); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		streams.Out,
		"config initialized at %s",
		c.ConfigPath,
	); err != nil {
		return err
	}
	return nil
}
