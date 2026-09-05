package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLegacyServicesTable(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.toml")
	legacyConfig := `[servers.legacy]
fqdn = "legacy.example.com"
ip = "192.168.1.10"
port = 443
enabled = true
`

	if err := os.WriteFile(configPath, []byte(legacyConfig), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	service, exists := cfg.Services["legacy"]
	if !exists {
		t.Fatal("legacy service was not loaded")
	}
	if service.FQDN != "legacy.example.com" {
		t.Errorf("service FQDN = %q, want %q", service.FQDN, "legacy.example.com")
	}
}
