package config

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadHTTPSettings(t *testing.T) {
	for _, tt := range []struct {
		name, fields, wantError string
	}{
		{"defaults", "", ""},
		{"expected error response", "expect_status = 404\nhttp_url = \"http://[::1]:8080/health?ready=1\"", ""},
		{"lower boundary", "expect_status = 100", ""},
		{"upper boundary", "expect_status = 599", ""},
		{"below range", "expect_status = 99", "expect_status"},
		{"above range", "expect_status = 600", "expect_status"},
		{"negative status", "expect_status = -1", "expect_status"},
		{"missing scheme", "http_url = \"example.com/health\"", "http_url"},
		{"unsupported scheme", "http_url = \"ftp://example.com\"", "http_url"},
		{"missing host", "http_url = \"https://\"", "http_url"},
		{"invalid port", "http_url = \"https://example.com:65536\"", "http_url"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.toml")
			content := "[services.web]\nfqdn = \"example.com\"\nip = \"127.0.0.1\"\nport = 443\n" + tt.fields
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := Load(path)
			if tt.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantError) || !strings.Contains(err.Error(), `service "web"`) {
				t.Fatalf("Load() error = %v, want service and field %q", err, tt.wantError)
			}
		})
	}
}
