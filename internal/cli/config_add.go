package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

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

func (c *ConfigAddCmd) Run(ctx context.Context, streams IOStreams) error {

	cfg, err := config.Load(c.ConfigPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if cfg.Servers == nil {
		cfg.Servers = make(map[string]config.Server)
	}

	if _, exists := cfg.Servers[c.ServerName]; exists {
		return fmt.Errorf("server %q already exists", c.ServerName)
	}

	newServer, err := promptServer(streams.In, streams.Out, c.ServerName)
	if err != nil {
		return err
	}

	cfg.Servers[c.ServerName] = newServer

	if err := config.ConfigAdd(c.ServerName, c.ConfigPath); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(
		streams.Out,
		"%s added to config",
		c.ServerName,
	); err != nil {
		return err
	}
	return nil
}

func promptServer(
	in io.Reader,
	out io.Writer,
	name string,
) (config.Server, error) {
	newServer := config.Server{
		Enabled: true,
	}
	if _, err := fmt.Fprintf(
		out,
		"Enter the fqdn for %s: ",
		name,
	); err != nil {
		return newServer,
			fmt.Errorf(
				"write fqdn: %w",
				err,
			)
	}
	if _, err := fmt.Fscan(in, &newServer.FQDN); err != nil {
		return newServer, fmt.Errorf("read FQDN: %w", err)
	}
	if _, err := fmt.Fprintf(
		out,
		"Enter %s's IP: ",
		name,
	); err != nil {
		return newServer,
			fmt.Errorf(
				"write ip: %w",
				err,
			)
	}
	if _, err := fmt.Fscan(in, &newServer.IP); err != nil {
		return newServer, fmt.Errorf("read ip address: %w", err)
	}
	if _, err := fmt.Fprintf(
		out,
		"Enter a port number for %s: ",
		name,
	); err != nil {
		return newServer,
			fmt.Errorf(
				"write port: %w",
				err,
			)
	}
	if _, err := fmt.Fscan(in, &newServer.Port); err != nil {
		return newServer, fmt.Errorf("read port: %w", err)
	}

	if newServer.Port <= 0 || newServer.Port > 65535 {
		return newServer, fmt.Errorf("port must be between 1 and 65535")
	}

	return newServer, nil

}
