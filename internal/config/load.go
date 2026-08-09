package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

func Load(cfgPath string) (Config, error) {
	var cfg Config

	_, err := toml.DecodeFile(cfgPath, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("load config: %w", err)
	}

	return cfg, nil
}
