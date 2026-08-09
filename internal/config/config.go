package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Servers map[string]Server `toml:"servers"`
}

type Server struct {
	FQDN    string `toml:"fqdn"`
	IP      string `toml:"ip"`
	Port    int    `toml:"port"`
	Enabled bool   `toml:"enabled"`
}

func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("determine user config directory: %w", err)
	}

	return filepath.Join(configDir, "homelabctl", "config.toml"), nil
}
