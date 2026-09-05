package config

import (
	"bytes"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

func AddService(
	configPath string,
	serviceName string,
	service Service,
) error {
	cfg, err := Load(configPath)
	if err != nil {
		return err
	}

	if cfg.Services == nil {
		cfg.Services = make(map[string]Service)
	}

	if _, exists := cfg.Services[serviceName]; exists {
		return fmt.Errorf("service %q already exists", serviceName)
	}

	cfg.Services[serviceName] = service

	if _, err := validateCfg(cfg); err != nil {
		return fmt.Errorf("validate new service: %w", err)
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
