package config

import (
	"bytes"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

func ConfigAdd(serverName string, configPath string) error {
	var cfg Config

	_, err := toml.DecodeFile(configPath, &cfg)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if cfg.Servers == nil {
		cfg.Servers = make(map[string]Server)
	}

	if _, exists := cfg.Servers[serverName]; exists {
		return fmt.Errorf("server %q already exists", serverName)
	}

	newServer := Server{
		Enabled: true,
	}

	fmt.Printf("Enter the fqdn for %s: ", serverName)
	if _, err := fmt.Scan(&newServer.FQDN); err != nil {
		return fmt.Errorf("read FQDN: %w", err)
	}

	fmt.Printf("Enter %s's IP: ", serverName)
	if _, err := fmt.Scan(&newServer.IP); err != nil {
		return fmt.Errorf("read IP address: %w", err)
	}

	fmt.Printf("Enter a port number for %s: ", serverName)
	if _, err := fmt.Scan(&newServer.Port); err != nil {
		return fmt.Errorf("read port: %w", err)
	}

	if newServer.Port <= 0 || newServer.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	cfg.Servers[serverName] = newServer

	var output bytes.Buffer
	if err := toml.NewEncoder(&output).Encode(cfg); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	if err := os.WriteFile(configPath, output.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
