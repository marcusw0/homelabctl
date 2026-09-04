package config

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/marcusw0/homelabctl/internal/check"
)

func newTestConfig(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")

	if err := Initialize(path); err != nil {
		t.Fatalf("Initialize() error: %v", err)
	}

	return path
}

func validTestService() Service {
	return Service{
		FQDN:    "myservice.example.com",
		IP:      "192.168.5.13",
		Port:    443,
		Enabled: true,
		Runbook: "runbooks/myservice.md",
	}
}

func TestAddService(t *testing.T) {
	cfgPath := newTestConfig(t)
	want := validTestService()

	if err := AddService(cfgPath, "myservice", want); err != nil {
		t.Fatalf("AddService() error: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	got, exists := cfg.Services["myservice"]
	if !exists {
		t.Fatal(`service "myservice" was not added`)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("service: %#v, want: %#v", got, want)
	}
}

func TestAddServiceOptionalFieldsRoundTrip(t *testing.T) {
	cfgPath := newTestConfig(t)
	followRedirects := false
	tlsWarnBefore := 30 * 24 * time.Hour
	want := validTestService()
	want.Checks = []check.Kind{check.KindHTTP, check.KindTLS}
	want.HTTPURL = "https://myservice.example.com/health"
	want.ExpectedStatus = 204
	want.FollowRedirects = &followRedirects
	want.TLSWarnBefore = &tlsWarnBefore
	want.Timeout = 3 * time.Second

	if err := AddService(cfgPath, "myservice", want); err != nil {
		t.Fatalf("AddService() error: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if got := cfg.Services["myservice"]; !reflect.DeepEqual(got, want) {
		t.Errorf("service: %#v, want: %#v", got, want)
	}
}

func TestAddRejectsDup(t *testing.T) {
	cfgPath := newTestConfig(t)
	service := validTestService()

	if err := AddService(cfgPath, "myservice", service); err != nil {
		t.Fatalf("AddService() error: %v", err)
	}

	before, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	err = AddService(cfgPath, "myservice", service)
	if err == nil {
		t.Fatal("Duplicate service was added with AddService(), want error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}

	after, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(before, after) {
		t.Error("config changed after duplicate was rejected")
	}
}

func TestAddRejectsInvalidNoConfigChange(t *testing.T) {
	negativeDuration := -time.Second

	tests := []struct {
		name    string
		service Service
	}{
		{
			name: "invalid hostname",
			service: Service{
				FQDN: "bad_host",
				IP:   "192.168.5.13",
				Port: 443,
			},
		},
		{
			name: "invalid ip",
			service: Service{
				FQDN: "myservice.example.com",
				IP:   "not-an-ip",
				Port: 443,
			},
		},
		{
			name: "invalid port",
			service: Service{
				FQDN: "myservice.example.com",
				IP:   "192.168.5.13",
				Port: 0,
			},
		},
		{
			name: "negative timeout",
			service: Service{
				FQDN:    "myservice.example.com",
				IP:      "192.168.5.13",
				Port:    443,
				Timeout: negativeDuration,
			},
		},
		{
			name: "negative TLS warning duration",
			service: Service{
				FQDN:          "myservice.example.com",
				IP:            "192.168.5.13",
				Port:          443,
				TLSWarnBefore: &negativeDuration,
			},
		},
		{
			name: "unsupported check",
			service: Service{
				FQDN:   "myservice.example.com",
				IP:     "192.168.5.13",
				Port:   443,
				Checks: []check.Kind{"htpt"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgPath := newTestConfig(t)

			before, err := os.ReadFile(cfgPath)
			if err != nil {
				t.Fatal(err)
			}

			if err := AddService(cfgPath, "myservice", tt.service); err == nil {
				t.Fatal("AddService() error = nil, want validation error")
			}

			after, err := os.ReadFile(cfgPath)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(before, after) {
				t.Error("config changed after invalid service was rejected")
			}
		})
	}
}
