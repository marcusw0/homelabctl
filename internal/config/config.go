package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/BurntSushi/toml"
	"github.com/marcusw0/homelabctl/internal/check"
)

type Config struct {
	Servers map[string]Server `toml:"servers"`
}

type Server struct {
	FQDN    string `toml:"fqdn"`
	IP      string `toml:"ip"`
	Port    int    `toml:"port"`
	Enabled bool   `toml:"enabled"`
	Runbook string `toml:"runbook"`
}

func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("determine user config directory: %w", err)
	}

	return filepath.Join(configDir, "homelabctl", "config.toml"), nil
}

func Load(cfgPath string) (Config, error) {
	var cfg Config

	_, err := toml.DecodeFile(cfgPath, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("load config: %w", err)
	}

	return validateCfg(cfg)
}

func validateCfg(cfg Config) (Config, error) {
	if len(cfg.Servers) == 0 {
		return cfg, nil
	}
	var errs []error

	names := make([]string, 0, len(cfg.Servers))
	for name := range cfg.Servers {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		server := cfg.Servers[name]
		if err := check.ValidateHostname(server.FQDN); err != nil {
			errs = append(errs, fmt.Errorf(
				"server %q FQDN: %w",
				name,
				err,
			),
			)
		}
		if err := check.ValidateIP(server.IP); err != nil {
			errs = append(errs, fmt.Errorf(
				"server %q IP: %w",
				name,
				err,
			),
			)
		}
		if err := check.ValidatePort(server.Port); err != nil {
			errs = append(errs, fmt.Errorf(
				"server %q Port: %w",
				name,
				err,
			),
			)
		}
	}

	return cfg, errors.Join(errs...)
}
