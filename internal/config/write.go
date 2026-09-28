package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

func AddService(path, name string, service Service) error {
	cfg, err := Load(path)
	if err != nil {
		return err
	}

	if cfg.Services == nil {
		cfg.Services = make(map[string]Service)
	}

	if _, exists := cfg.Services[name]; exists {
		return fmt.Errorf("service %q already exists", name)
	}

	cfg.Services[name] = service

	if _, err := validateCfg(cfg); err != nil {
		return fmt.Errorf("validate new service: %w", err)
	}
	return writeConfig(path, cfg)
}

func UpdateService(path, name string, service Service) error {
	cfg, err := Load(path)
	if err != nil {
		return err
	}

	cfg.Services[name] = service

	if _, err := validateCfg(cfg); err != nil {
		return fmt.Errorf("validate new service: %w", err)
	}
	return writeConfig(path, cfg)
}

func writeConfig(path string, cfg Config) error {
	var output bytes.Buffer
	if err := toml.NewEncoder(&output).Encode(cfg); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".homelabctl-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(output.Bytes()); err != nil {
		tmp.Close()
		return fmt.Errorf("write tmp config: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync tmp config: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close tmp config: %w", err)
	}

	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}
	return nil
}
