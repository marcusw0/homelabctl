package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/marcusw0/homelabctl/internal/check"
)

type Config struct {
	Services map[string]Service `toml:"services"`
}

type Service struct {
	FQDN    string `toml:"fqdn"`
	IP      string `toml:"ip"`
	Port    int    `toml:"port"`
	Enabled bool   `toml:"enabled"`
	Runbook string `toml:"runbook"`

	Checks []check.Kind `toml:"checks,omitempty"`

	// Optional URL override. When empty, derive HTTPS URL from FQDN.
	HTTPURL         string `toml:"http_url,omitempty"`
	ExpectedStatus  int    `toml:"expect_status,omitempty"`
	FollowRedirects *bool  `toml:"follow_redirects,omitempty"`

	// Set when to be warned of an approaching certificate expiration.
	TLSWarnBefore *time.Duration `toml:"tls_warn_before,omitempty"`
	Timeout       time.Duration  `toml:"timeout,omitempty"`
}

func DefaultPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("determine user config directory: %w", err)
	}

	return filepath.Join(configDir, "homelabctl", "config.toml"), nil
}

func Load(cfgPath string) (Config, error) {
	var file struct {
		Services map[string]Service `toml:"services"`
		// LegacyServices keeps existing [servers] configurations readable.
		LegacyServices map[string]Service `toml:"servers"`
	}

	_, err := toml.DecodeFile(cfgPath, &file)
	if err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}

	cfg := Config{Services: file.Services}
	if len(file.LegacyServices) > 0 {
		if cfg.Services == nil {
			cfg.Services = make(map[string]Service, len(file.LegacyServices))
		}

		for name, service := range file.LegacyServices {
			if _, exists := cfg.Services[name]; exists {
				return Config{}, fmt.Errorf(
					"load config: service %q is defined in both [services] and legacy [servers] tables",
					name,
				)
			}
			cfg.Services[name] = service
		}
	}

	return validateCfg(cfg)
}

func validateCfg(cfg Config) (Config, error) {
	if len(cfg.Services) == 0 {
		return cfg, nil
	}
	var errs []error

	names := make([]string, 0, len(cfg.Services))
	for name := range cfg.Services {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		service := cfg.Services[name]
		if service.ExpectedStatus != 0 && (service.ExpectedStatus < 100 || service.ExpectedStatus > 599) {
			errs = append(errs, fmt.Errorf("service %q expect_status must be between 100 and 599 (or 0 for the default)", name))
		}
		if service.HTTPURL != "" {
			if !strings.Contains(service.HTTPURL, "://") {
				errs = append(errs, fmt.Errorf("service %q http_url must be a full HTTP or HTTPS URL", name))
			} else if normalized, err := check.NormalizeHTTPURL(service.HTTPURL); err != nil {
				errs = append(errs, fmt.Errorf("service %q http_url: %w", name, err))
			} else {
				service.HTTPURL = normalized
				cfg.Services[name] = service
			}
		}
		for _, kind := range service.Checks {
			if !kind.Valid() {
				errs = append(errs, fmt.Errorf(
					"service %q has unsupported check %q",
					name,
					kind,
				))
			}
		}
		if err := check.ValidateHostname(service.FQDN); err != nil {
			errs = append(errs, fmt.Errorf(
				"service %q FQDN: %w",
				name,
				err,
			),
			)
		}
		if err := check.ValidateIP(service.IP); err != nil {
			errs = append(errs, fmt.Errorf(
				"service %q IP: %w",
				name,
				err,
			),
			)
		}
		if err := check.ValidatePort(service.Port); err != nil {
			errs = append(errs, fmt.Errorf(
				"service %q Port: %w",
				name,
				err,
			),
			)
		}
		if service.Timeout < 0 {
			errs = append(errs, fmt.Errorf(
				"service %q timeout must not be negative",
				name,
			))
		}
		if service.TLSWarnBefore != nil && *service.TLSWarnBefore < 0 {
			errs = append(errs, fmt.Errorf(
				"service %q TLS warning duration must not be negative",
				name,
			))
		}
	}

	return cfg, errors.Join(errs...)
}
