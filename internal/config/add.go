package config

import (
	"bytes"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

func AddServer(
	configPath string,
	serverName string,
	server Server,
) error {
	cfg, err := Load(configPath)
	if err != nil {
		return err
	}

	if cfg.Servers == nil {
		cfg.Servers = make(map[string]Server)
	}

	if _, exists := cfg.Servers[serverName]; exists {
		return fmt.Errorf("server %q already exists", serverName)
	}

	cfg.Servers[serverName] = server

	if _, err := validateCfg(cfg); err != nil {
		return fmt.Errorf("validate new server: %w", err)
	}

	var output bytes.Buffer
	if err := toml.NewEncoder(&output).Encode(cfg); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	if err := os.WriteFile(configPath, output.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
